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

type ShellyHT struct {
	Method string `json:"method"`
	Src    string `json:"src"`
	Dst    string `json:"dst"`
	Params struct {
		WS struct {
			Connected bool `json:"connected"`
		} `json:"ws"`
		MQTT struct {
			Connected bool `json:"connected"`
		} `json:"mqtt"`
		Cloud struct {
			Connected bool `json:"connected"`
		} `json:"cloud"`
		HTUI interface{} `json:"ht_ui"`
		Sys  struct {
			RAMSize      int         `json:"ram_size"`
			RAMFree      int         `json:"ram_free"`
			CfgRev       int         `json:"cfg_rev"`
			Time         interface{} `json:"time"`
			ResetReason  int         `json:"reset_reason"`
			WebhookRev   int         `json:"webhook_rev"`
			FSSize       int         `json:"fs_size"`
			MAC          string      `json:"mac"`
			WakeupReason struct {
				Boot  string `json:"boot"`
				Cause string `json:"cause"`
			} `json:"wakeup_reason"`
			Unixtime         interface{}            `json:"unixtime"`
			FSFree           int                    `json:"fs_free"`
			AvailableUpdates map[string]interface{} `json:"available_updates"`
			Uptime           int                    `json:"uptime"`
			KvsRev           int                    `json:"kvs_rev"`
			RestartRequired  bool                   `json:"restart_required"`
			WakeupPeriod     int                    `json:"wakeup_period"`
		} `json:"sys"`
		DevicePower struct {
			External struct {
				Present bool `json:"present"`
			} `json:"external"`
			Battery struct {
				Percent int     `json:"percent"`
				V       float64 `json:"V"`
			} `json:"battery"`
			ID int `json:"id"`
		} `json:"devicepower:0"`
		BLE      struct{} `json:"ble"`
		TS       float64  `json:"ts"`
		Humidity struct {
			RH float64 `json:"rh"`
			ID int     `json:"id"`
		} `json:"humidity:0"`
		WiFi struct {
			Status string `json:"status"`
			StaIP  string `json:"sta_ip"`
			RSSI   int    `json:"rssi"`
			SSID   string `json:"ssid"`
		} `json:"wifi"`
		Temperature struct {
			TC float64 `json:"tC"`
			TF float64 `json:"tF"`
			ID int     `json:"id"`
		} `json:"temperature:0"`
	} `json:"params"`
}
