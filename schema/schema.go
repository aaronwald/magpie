package schema

type NodeUpdate struct {
	Node        string `json:"device"`
	Timestamp   string `json:"timestamp"`
	Description string `json:"description"`
}

type ShellyMotion struct {
	Encryption    bool   `json:"encryption"`
	BTHomeVersion int    `json:"BTHome_version"`
	Pid           int    `json:"pid"`
	Battery       int    `json:"Battery"`
	Illuminance   int    `json:"Illuminance"`
	Motion        int    `json:"Motion"`
	Addr          string `json:"addr"`
	Rssi          int    `json:"rssi"`
}

// {\"encryption\":false,\"BTHome_version\":2,\"pid\":124,\"Battery\":100,\"Illuminance\":10,\"Window\":0,\"Rotation\":0,\"addr\":\"b0:c7:de:2c:20:0e\",\"rssi\":-79}"

type ShellyOpenClose struct {
	Encryption    bool   `json:"encryption"`
	BTHomeVersion int    `json:"BTHome_version"`
	Pid           int    `json:"pid"`
	Battery       int    `json:"Battery"`
	Illuminance   int    `json:"Illuminance"`
	Window        int    `json:"Window"`
	Rotation      int    `json:"Rotation"`
	Addr          string `json:"addr"`
	Rssi          int    `json:"rssi"`
}

type ShellyRpcInput struct {
	Src    string `json:"src"`
	Dst    string `json:"dst"`
	Method string `json:"method"`
	Params struct {
		Timestamp float64 `json:"ts"`
		Input     struct {
			Id    int  `json:"id"`
			State bool `json:"state"`
		} `json:"input:0"`
		Switch struct {
			Id             int     `json:"id"`
			Output         bool    `json:"output"`
			Source         string  `json:"source"`
			TimerDuration  float32 `json:"timer_duration"`
			TimerStartedAt float64 `json:"timer_started_at"`
		} `json:"switch:0"`
	}
}
