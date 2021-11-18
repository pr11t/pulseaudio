package pulseaudio

import "io"

// Sink contains information about a sink in pulseaudio
type Sink struct {
	Index              uint32
	Name               string
	Description        string
	SampleSpec         sampleSpec
	ChannelMap         channelMap
	ModuleIndex        uint32
	Cvolume            cvolume
	Muted              bool
	MonitorSourceIndex uint32
	MonitorSourceName  string
	Latency            uint64
	Driver             string
	Flags              uint32
	PropList           map[string]string
	RequestedLatency   uint64
	BaseVolume         uint32
	SinkState          uint32
	NVolumeSteps       uint32
	CardIndex          uint32
	Ports              []sinkPort
	ActivePortName     string
	Formats            []formatInfo
}

// ReadFrom deserializes a sink packet from pulseaudio
func (s *Sink) ReadFrom(r io.Reader) (int64, error) {
	var portCount uint32
	err := bread(r,
		uint32Tag, &s.Index,
		stringTag, &s.Name,
		stringTag, &s.Description,
		&s.SampleSpec,
		&s.ChannelMap,
		uint32Tag, &s.ModuleIndex,
		&s.Cvolume,
		&s.Muted,
		uint32Tag, &s.MonitorSourceIndex,
		stringTag, &s.MonitorSourceName,
		usecTag, &s.Latency,
		stringTag, &s.Driver,
		uint32Tag, &s.Flags,
		&s.PropList,
		usecTag, &s.RequestedLatency,
		volumeTag, &s.BaseVolume,
		uint32Tag, &s.SinkState,
		uint32Tag, &s.NVolumeSteps,
		uint32Tag, &s.CardIndex,
		uint32Tag, &portCount)
	if err != nil {
		return 0, err
	}
	s.Ports = make([]sinkPort, portCount)
	for i := uint32(0); i < portCount; i++ {
		err = bread(r, &s.Ports[i])
		if err != nil {
			return 0, err
		}
	}
	if portCount == 0 {
		err = bread(r, stringNullTag)
		if err != nil {
			return 0, err
		}
	} else {
		err = bread(r, stringTag, &s.ActivePortName)
		if err != nil {
			return 0, err
		}
	}

	var formatCount uint8
	err = bread(r,
		uint8Tag, &formatCount)
	if err != nil {
		return 0, err
	}
	s.Formats = make([]formatInfo, formatCount)
	for i := uint8(0); i < formatCount; i++ {
		err = bread(r, &s.Formats[i])
		if err != nil {
			return 0, err
		}
	}
	return 0, nil
}

// Sinks queries PulseAudio for a list of sinks and returns an array
func (c *Client) Sinks() ([]Sink, error) {
	b, err := c.request(commandGetSinkInfoList)
	if err != nil {
		return nil, err
	}
	var sinks []Sink
	for b.Len() > 0 {
		var sink Sink
		err = bread(b, &sink)
		if err != nil {
			return nil, err
		}
		sinks = append(sinks, sink)
	}
	return sinks, nil
}

type sinkPort struct {
	Name, Description string
	Pririty           uint32
	Available         uint32
}

func (p *sinkPort) ReadFrom(r io.Reader) (int64, error) {
	return 0, bread(r,
		stringTag, &p.Name,
		stringTag, &p.Description,
		uint32Tag, &p.Pririty,
		uint32Tag, &p.Available)
}

func (c *Client) SetDefaultSink(sinkName string) error {
	_, err := c.request(commandSetDefaultSink,
		stringTag, []byte(sinkName), byte(0))
	return err
}

func (c *Client) SetSinkPort(sinkName, portName string) error {
	_, err := c.request(commandSetSinkPort,
		uint32Tag, uint32(0xffffffff),
		stringTag, []byte(sinkName),
		byte(0),
		stringTag, []byte(portName),
		byte(0))
	return err
}

type SinkInput struct {
	Index          uint32
	Name           string
	OwnerModule    uint32
	Client         uint32
	Sink           uint32
	SampleSpec     sampleSpec
	ChannelMap     channelMap
	Cvolume        cvolume
	BufferUsec     uint64
	SinkUsec       uint64
	ResampleMethod string
	Driver         string
	Mute           bool
	PropList       map[string]string
	Corked         bool
	HasVolume      bool
	VolumeWritable bool
	Formats        []formatInfo
}

func (s *SinkInput) ReadFrom(r io.Reader) (int64, error) {
	err := bread(r,
		uint32Tag, &s.Index,
		stringTag, &s.Name,
		uint32Tag, &s.OwnerModule,
		uint32Tag, &s.Client,
		uint32Tag, &s.Sink,
		&s.SampleSpec,
		&s.ChannelMap,
		&s.Cvolume,
		usecTag, &s.BufferUsec,
		usecTag, &s.SinkUsec,
		stringTag, &s.ResampleMethod,
		stringTag, &s.Driver,
		&s.Mute,
		&s.PropList,
		&s.Corked,
		&s.HasVolume,
		&s.VolumeWritable)
	if err != nil {
		return 0, err
	}
	fi := formatInfo{}

	var formatCount uint8
	err = bread(r, formatInfoTag,
		uint8Tag, &formatCount)
	if err != nil {
		return 0, err
	}

	err = bread(r, &fi.PropList)
	return 0, nil
}

func (c *Client) SinkInputs() ([]SinkInput, error) {
	b, err := c.request(commandGetSinkInputInfoList)
	if err != nil {
		return nil, err
	}
	var sinkInputs []SinkInput
	for b.Len() > 0 {
		var sinkInputInfo SinkInput
		err = bread(b, &sinkInputInfo)
		if err != nil {
			return nil, err
		}
		sinkInputs = append(sinkInputs, sinkInputInfo)
	}
	return sinkInputs, nil
}

func (c *Client) MoveSinkInput(inputIndex uint32, sinkName string) error {
	_, err := c.request(commandMoveSinkInput,
		uint32Tag, uint32(inputIndex),
		uint32Tag, uint32(0xffffffff),
		stringTag, []byte(sinkName),
		byte(0))
	return err
}
