// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package profileuiexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	// "go.opentelemetry.io/collector/pdata/pcommon" // Temporarily removed
	"go.opentelemetry.io/collector/pdata/pprofile"
	"go.uber.org/zap"
)

// TimeSeriesDataPoint represents a single point in our time series data
type TimeSeriesDataPoint struct {
	Timestamp          time.Time        `json:"timestamp"`
	SampleCountsByType map[string]int64 `json:"sampleCountsByType"`
}

// profilesExporter is the exporter that displays profiles in a UI.
type profilesExporter struct {
	config               *Config
	logger               *zap.Logger
	server               *http.Server
	mu                   sync.Mutex // Protects timeSeries and currentBucketSamples
	timeSeries           []TimeSeriesDataPoint
	currentBucketSamples map[string]int64 // Changed to map
	aggregationTicker    *time.Ticker
	shutdownTickerChan   chan struct{}
}

// newProfilesExporter creates a new instance of the profilesExporter.
func newProfilesExporter(cfg *Config, set exporter.Settings) (*profilesExporter, error) {
	pe := &profilesExporter{
		config:               cfg,
		logger:               set.Logger,
		timeSeries:           make([]TimeSeriesDataPoint, 0),
		currentBucketSamples: make(map[string]int64), // Initialize as an empty map
		shutdownTickerChan:   make(chan struct{}),
	}
	return pe, nil
}

// Start is called when the exporter is started. It launches the HTTP server.
func (pe *profilesExporter) Start(_ context.Context, host component.Host) error {
	pe.logger.Info("Starting Profile UI exporter", zap.Int("port", pe.config.HTTPPort))
	mux := http.NewServeMux()
	mux.HandleFunc("/", pe.handleViewProfilesPage) // Serves the HTML page
	mux.HandleFunc("/data", pe.handleDataEndpoint) // Serves the time-series data

	pe.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", pe.config.HTTPPort),
		Handler: mux,
	}

	// Start the HTTP server in a goroutine
	go func() {
		if err := pe.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pe.logger.Error("Profile UI server failed to start", zap.Error(err))
		}
	}()

	// Start the aggregation ticker
	pe.aggregationTicker = time.NewTicker(5 * time.Second) // Aggregate every 5 seconds
	go func() {
		for {
			select {
			case <-pe.aggregationTicker.C:
				pe.mu.Lock()
				now := time.Now()
				currentSamplesCopy := make(map[string]int64)
				for k, v := range pe.currentBucketSamples {
					currentSamplesCopy[k] = v
				}
				pe.timeSeries = append(pe.timeSeries, TimeSeriesDataPoint{Timestamp: now, SampleCountsByType: currentSamplesCopy})
				pe.currentBucketSamples = make(map[string]int64) // Reset for the new intervals to prevent unbounded growth
				// if len(pe.timeSeries) > 1000 { // Keep, for example, last 1000 points
				// 	pe.timeSeries = pe.timeSeries[len(pe.timeSeries)-1000:]
				// }
				pe.mu.Unlock()
			case <-pe.shutdownTickerChan:
				return
			}
		}
	}()

	return nil
}

// Shutdown is called when the exporter is stopped. It gracefully shuts down the HTTP server.
func (pe *profilesExporter) Shutdown(ctx context.Context) error {
	pe.logger.Info("Shutting down Profile UI exporter")

	if pe.aggregationTicker != nil {
		pe.aggregationTicker.Stop()
	}
	close(pe.shutdownTickerChan) // Signal the ticker goroutine to stop

	if pe.server != nil {
		return pe.server.Shutdown(ctx)
	}
	return nil
}

// consumeProfiles is called by the collector pipeline with profile data.
func (pe *profilesExporter) consumeProfiles(_ context.Context, pd pprofile.Profiles) error {
	batchCountsByType := make(map[string]int64)

	dictionary := pd.ProfilesDictionary()
	attributeTable := dictionary.AttributeTable()

	rps := pd.ResourceProfiles()
	for _, rs := range rps.All() {
		for _, sp := range rs.ScopeProfiles().All() {
			for _, profile := range sp.Profiles().All() {
				locationIndices := profile.LocationIndices()

				// Iterate over all samples in this profile.
				for _, sample := range profile.Sample().All() {
					// Iterate over all locations in this sample.
					for lIdx := range sample.LocationsLength() {
						locTableIdx := locationIndices.At(int(sample.LocationsStartIndex() + lIdx))
						location := dictionary.LocationTable().At(int(locTableIdx))

						// Iterate over all attributes in this location.
						var foundSampleTypeInLocation bool
						for _, attributeIndex := range location.AttributeIndices().All() {
							attribute := attributeTable.At(int(attributeIndex))
							keyString := attribute.Key()
							pe.logger.Debug("Location attribute", zap.String("key", keyString))

							if keyString == "profile.frame.type" {
								attrVal := attribute.Value()
								pe.logger.Debug("Attribute value details", zap.String("type", attrVal.Type().String()), zap.String("as_string", attrVal.AsString()), zap.Int64("as_int", attrVal.Int()))

								// Corrected: Attribute value is directly a string.
								valueString := attrVal.Str()
								pe.logger.Info("Found profile.frame.type", zap.String("type", valueString))
								batchCountsByType[valueString]++
								foundSampleTypeInLocation = true
								break
							}
						}
						if !foundSampleTypeInLocation {
							// Optionally log if a location didn't have the specific attribute
							// pe.logger.Debug("profile.sample.type not found for a location")
						}
					}
				}
			}
		}
	}

	pe.logger.Debug("Final batchCountsByType before lock", zap.Any("counts", batchCountsByType))

	pe.mu.Lock()
	for sType, count := range batchCountsByType {
		pe.currentBucketSamples[sType] += count
	}
	pe.mu.Unlock()

	pe.logger.Debug("Received profiles", zap.Any("current_bucket_totals_by_type", pe.currentBucketSamples))
	return nil
}

// handleViewProfilesPage serves the main HTML page for the UI.
func (pe *profilesExporter) handleViewProfilesPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Basic HTML structure. We'll enhance this later with JavaScript for charting.
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <title>Profile UI Exporter</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/date-fns@^2/index.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chartjs-adapter-date-fns@^3/dist/chartjs-adapter-date-fns.bundle.min.js"></script>
    <style>
        body { font-family: sans-serif; margin: 20px; }
        #chart-container { width: 80%; max-width: 900px; margin-bottom: 30px; }
        table { border-collapse: collapse; width: 80%; max-width: 900px; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <h1>Profile UI Exporter - Samples Over Time</h1>
    <div id="chart-container" style="height: 70vh; width: 90vw; margin: auto;">
        <canvas id="profileChart"></canvas>
    </div>
    
    <script>
      function initializeApp() {
        console.log("Attempting to initialize app. Current document.body.innerHTML:");
        console.log(document.body.innerHTML);
        const canvasElement = document.getElementById('profileChart');
        if (!canvasElement) {
            console.error("'profileChart' canvas element not found in the DOM.");
            return;
        }
        const ctx = canvasElement.getContext('2d');
        if (!ctx) {
            console.error("Failed to get 2D context from 'profileChart' canvas element.");
            return;
        }

        // Make profileChart a global or appropriately scoped variable if updateChart/fetchData are outside initializeApp
        // For this structure, it's scoped to initializeApp, so helper functions must also be inside or passed the chart instance.
        const profileChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [], // Timestamps
                datasets: [] // Datasets will be added dynamically
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    x: {
                        type: 'time',
                        time: {
                            unit: 'second',
                            displayFormats: { 
                                second: 'HH:mm:ss'
                            }
                        },
                        title: {
                            display: true,
                            text: 'Time'
                        }
                    },
                    y: {
                        title: {
                            display: true,
                            text: 'Sample Count'
                        },
                        beginAtZero: true
                    }
                },
                animation: {
                    duration: 0 
                },
                plugins: {
                    legend: {
                        position: 'top',
                    }
                }
            }
        });

        const colorPalette = [
            'rgb(255, 99, 132)', 'rgb(54, 162, 235)', 'rgb(255, 206, 86)',
            'rgb(75, 192, 192)', 'rgb(153, 102, 255)', 'rgb(255, 159, 64)',
            'rgb(201, 203, 207)', 'rgb(0, 0, 0)'
        ];
        let colorIndex = 0;

        function updateChart(newData) {
            // No need to check for profileChart here as it's in the same scope and guaranteed to exist if this function is called after its init.
            if (!newData || newData.length === 0) {
                profileChart.data.labels = [];
                profileChart.data.datasets = [];
                profileChart.update();
                return;
            }

            const labels = newData.map(dp => dp.timestamp);
            profileChart.data.labels = labels;

            const allKeys = new Set();
            newData.forEach(dp => {
                if (dp.sampleCountsByType) {
                    Object.keys(dp.sampleCountsByType).forEach(key => allKeys.add(key));
                }
            });
            
            let currentDatasets = profileChart.data.datasets;
            
            allKeys.forEach(typeKey => {
                let dataset = currentDatasets.find(ds => ds.label === typeKey);
                if (!dataset) {
                    dataset = {
                        label: typeKey,
                        borderColor: colorPalette[colorIndex % colorPalette.length],
                        tension: 0.1,
                        fill: false,
                        data: []
                    };
                    currentDatasets.push(dataset);
                    colorIndex++;
                }
                dataset.data = newData.map(dp => (dp.sampleCountsByType && dp.sampleCountsByType[typeKey]) || 0);
            });
            
            profileChart.data.datasets = currentDatasets.filter(ds => allKeys.has(ds.label));
            profileChart.update();
        }

        async function fetchData() {
            try {
                const response = await fetch('/data');
                if (!response.ok) {
                    console.error('Error fetching data: ' + response.statusText);
                    // updateTable([]); // Removed call
                    updateChart([]); // Clear chart on error
                    return;
                }
                const data = await response.json();
                updateChart(data);
                // updateTable(data); // Removed call
            } catch (error) {
                console.error('Error fetching or processing data:', error);
                // updateTable([]); // Removed call
                updateChart([]); // Clear chart on error
            }
        }
        fetchData(); 
        setInterval(fetchData, 5000);
      } // End of initializeApp

      // Use window.onload to ensure all resources are loaded
      window.addEventListener('load', initializeApp);
    </script>
`
	fmt.Fprint(w, htmlContent)
}

// ... (rest of the code remains the same)
func (pe *profilesExporter) handleDataEndpoint(w http.ResponseWriter, r *http.Request) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Return a copy to avoid race conditions if the slice is modified while marshalling
	dataToServe := make([]TimeSeriesDataPoint, len(pe.timeSeries))
	copy(dataToServe, pe.timeSeries)

	if err := json.NewEncoder(w).Encode(dataToServe); err != nil {
		pe.logger.Error("Failed to encode time-series data to JSON", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
