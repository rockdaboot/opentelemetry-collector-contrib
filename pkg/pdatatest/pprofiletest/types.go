// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package pprofiletest // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pprofiletest"
import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pprofile"
)

type Profiles struct {
	ResourceProfiles []ResourceProfile
}

func (p Profiles) Transform() pprofile.Profiles {
	pp := pprofile.NewProfiles()
	for _, rp := range p.ResourceProfiles {
		rp.Transform(pp)
	}
	return pp
}

type ResourceProfile struct {
	ScopeProfiles []ScopeProfile
	Resource      Resource
}

func (rp ResourceProfile) Transform(pp pprofile.Profiles) pprofile.ResourceProfiles {
	prp := pp.ResourceProfiles().AppendEmpty()
	for _, sp := range rp.ScopeProfiles {
		sp.Transform(prp)
	}
	for k, v := range rp.Resource.Attributes {
		prp.Resource().Attributes().PutStr(k, v)
	}
	return prp
}

type Resource struct {
	Attributes map[string]string
}

type ScopeProfile struct {
	Profile []Profile
}

func (sp ScopeProfile) Transform(prp pprofile.ResourceProfiles) pprofile.ScopeProfiles {
	psp := prp.ScopeProfiles().AppendEmpty()
	for _, p := range sp.Profile {
		p.Transform(psp)
	}
	return psp
}

type Profile struct {
	sampleType             []ValueType
	sample                 []Sample
	timeNanos              pcommon.Timestamp
	durationNanos          pcommon.Timestamp
	periodType             ValueType
	period                 int64
	comment                []string
	defaultSampleType      ValueType
	profileID              pprofile.ProfileID
	droppedAttributesCount uint32
	originalPayloadFormat  string
	originalPayload        []byte
	attributes             []Attribute
}

func (p *Profile) Transform(psp pprofile.ScopeProfiles) pprofile.Profile {
	pp := psp.Profiles().AppendEmpty()
	for _, st := range p.sampleType {
		st.Transform(pp)
	}
	for _, sa := range p.sample {
		sa.Transform(pp)
	}
	pp.SetTime(p.timeNanos)
	pp.SetDuration(p.durationNanos)
	p.periodType.Transform(pp)
	pp.SetPeriod(p.period)
	for _, c := range p.comment {
		pp.CommentStrindices().Append(addString(pp, c))
	}
	p.defaultSampleType.Transform(pp)
	pp.SetProfileID(p.profileID)
	pp.SetDroppedAttributesCount(p.droppedAttributesCount)
	pp.SetOriginalPayloadFormat(p.originalPayloadFormat)
	pp.OriginalPayload().FromRaw(p.originalPayload)
	for _, at := range p.attributes {
		pp.AttributeIndices().Append(at.Transform(pp))
	}

	return pp
}

func addString(pp pprofile.Profile, s string) int32 {
	for i := range pp.StringTable().Len() {
		if pp.StringTable().At(i) == s {
			return int32(i)
		}
	}
	pp.StringTable().Append(s)
	return int32(pp.StringTable().Len() - 1)
}

type ValueType struct {
	typ                    string
	unit                   string
	aggregationTemporality pprofile.AggregationTemporality
}

func (vt *ValueType) Transform(pp pprofile.Profile) {
	pvt := pp.SampleType().AppendEmpty()
	pvt.SetTypeStrindex(addString(pp, vt.typ))
	pvt.SetUnitStrindex(addString(pp, vt.unit))
	pvt.SetAggregationTemporality(vt.aggregationTemporality)
}

type Sample struct {
	link               *Link // optional
	value              []int64
	locations          []Location
	attributes         []Attribute
	timestampsUnixNano []uint64
}

func (sa *Sample) Transform(pp pprofile.Profile) {
	if len(sa.value) != pp.SampleType().Len() {
		panic("length of profile.sample_type must be equal to the length of sample.value")
	}
	psa := pp.Sample().AppendEmpty()
	psa.SetLocationsStartIndex(int32(pp.LocationTable().Len()))
	for _, loc := range sa.locations {
		ploc := pp.LocationTable().AppendEmpty()
		if loc.mapping != nil {
			loc.mapping.Transform(pp)
		}
		ploc.SetAddress(loc.address)
		ploc.SetIsFolded(loc.isFolded)
		for _, l := range loc.line {
			pl := ploc.Line().AppendEmpty()
			pl.SetLine(l.line)
			pl.SetColumn(l.column)
			pl.SetFunctionIndex(l.function.Transform(pp))
		}
		for _, at := range loc.attributes {
			ploc.AttributeIndices().Append(at.Transform(pp))
		}
	}
	psa.SetLocationsLength(int32(pp.LocationTable().Len()) - psa.LocationsStartIndex())
	psa.Value().FromRaw(sa.value)
	for _, at := range sa.attributes {
		psa.AttributeIndices().Append(at.Transform(pp))
	}
	//nolint:revive,staticcheck
	if sa.link != nil {
		// psa.SetLinkIndex(sa.link.Transform(pp)) <-- undefined yet
	}
	psa.TimestampsUnixNano().FromRaw(sa.timestampsUnixNano)
}

type Location struct {
	mapping    *Mapping
	address    uint64
	line       []Line
	isFolded   bool
	attributes []Attribute
}

type Link struct {
	traceID pcommon.TraceID
	spanID  pcommon.SpanID
}

func (l *Link) Transform(pp pprofile.Profile) int32 {
	pl := pp.LinkTable().AppendEmpty()
	pl.SetTraceID(l.traceID)
	pl.SetSpanID(l.spanID)
	return int32(pp.LinkTable().Len() - 1)
}

type Mapping struct {
	memoryStart     uint64
	memoryLimit     uint64
	fileOffset      uint64
	filename        string
	attributes      []Attribute
	hasFunctions    bool
	hasFileNames    bool
	hasLineNumbers  bool
	hasInlineFrames bool
}

func (m *Mapping) Transform(pp pprofile.Profile) {
	pm := pp.MappingTable().AppendEmpty()
	pm.SetMemoryStart(m.memoryStart)
	pm.SetMemoryLimit(m.memoryLimit)
	pm.SetFileOffset(m.fileOffset)
	pm.SetFilenameStrindex(addString(pp, m.filename))
	for _, at := range m.attributes {
		pm.AttributeIndices().Append(at.Transform(pp))
	}
	pm.SetHasFunctions(m.hasFunctions)
	pm.SetHasFilenames(m.hasFileNames)
	pm.SetHasLineNumbers(m.hasLineNumbers)
	pm.SetHasInlineFrames(m.hasInlineFrames)
}

type Attribute struct {
	key   string
	value string
}

func (a *Attribute) Transform(pp pprofile.Profile) int32 {
	pa := pp.AttributeTable().AppendEmpty()
	pa.SetKey(a.key)
	pa.Value().SetStr(a.value)
	return int32(pp.AttributeTable().Len() - 1)
}

type Line struct {
	line     int64
	column   int64
	function Function
}

type Function struct {
	name       string
	systemName string
	filename   string
	startLine  int64
}

func (f *Function) Transform(pp pprofile.Profile) int32 {
	pf := pp.FunctionTable().AppendEmpty()
	pf.SetNameStrindex(addString(pp, f.name))
	pf.SetSystemNameStrindex(addString(pp, f.systemName))
	pf.SetFilenameStrindex(addString(pp, f.filename))
	pf.SetStartLine(f.startLine)
	return int32(pp.FunctionTable().Len() - 1)
}
