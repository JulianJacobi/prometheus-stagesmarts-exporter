package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "strconv"

    prom "github.com/prometheus/client_golang/prometheus"
)

type PduBool struct {
    Value bool
}

func (pb *PduBool) UnmarshalJSON(data []byte) error {
    var s string
    err := json.Unmarshal(data, &s)
    if err != nil {
        return err
    }
    if s == "true" {
        pb.Value = true
    } else if s == "false" {
        pb.Value = false
    } else {
        return errors.New("Unexpected string value, expext true or false")
    }
    return nil
}

func (pb *PduBool) ToFloat64() float64 {
    if pb.Value {
        return 1
    } else {
        return 0
    }
}

type PduFloat64 struct {
    Value float64
}

func (pf *PduFloat64) UnmarshalJSON(data []byte) error {
    var s string
    err := json.Unmarshal(data, &s)
    if err != nil {
        return err
    }
    f, err := strconv.ParseFloat(s, 64)
    pf.Value = f
    return err
}

type SingletonList[T any] struct {
    Value T
}

func (sl *SingletonList[T]) UnmarshalJSON(data []byte) error {
    var l []T
    err := json.Unmarshal(data, &l)
    if err != nil {
        return err
    }
    if len(l) != 1 {
        return errors.New(fmt.Sprintf("Expected list with single value, got %d values.", len(l)))
    }
    sl.Value = l[0]
    return nil
}

type PduValueResponse struct {
    PDU PduValues `json:"smartPDU"`
}

func (pvr *PduValueResponse) GetMetrics() []prom.Collector {
    return pvr.PDU.GetMetrics()
}

type PduValues struct {
    OnBattery       SingletonList[PduBool]         `json:"onBattery"`
    MainInputValues SingletonList[MainInputValues] `json:"mainInputValues"`
    Temparature     SingletonList[Temperature]     `json:"temperature"`
    Breakers        SingletonList[Breakers]        `json:"breakers"`
    OutputChannels  SingletonList[OutputChannels]  `json:"outputChannels"`
    Config          Config                         `json:"config"`
}

func (pv *PduValues) GetMetrics() []prom.Collector {
    onBattery := prom.NewGaugeVec(
        prom.GaugeOpts{
            Namespace: "stagesmarts",
            Subsystem: "system",
            Name: "on_battery",
            Help: "System on battery",
        },
        []string{},
    )
    onBattery.WithLabelValues().Set(pv.OnBattery.Value.ToFloat64())
    ret := []prom.Collector{
        onBattery,
    }
    ret = append(ret, pv.MainInputValues.Value.GetMetrics()...)
    ret = append(ret, pv.Temparature.Value.GetMetrics()...)
    ret = append(ret, pv.Breakers.Value.GetMetrics()...)
    ret = append(ret, pv.OutputChannels.Value.GetMetrics(&pv.Config)...)
    ret = append(ret, pv.Config.GetMetrics()...)
    return ret
}

type MainInputValues struct {
    L1Voltage   SingletonList[PduFloat64] `json:"L1V"`
    L2Voltage   SingletonList[PduFloat64] `json:"L2V"`
    L3Voltage   SingletonList[PduFloat64] `json:"L3V"`
    L1L2Voltage SingletonList[PduFloat64] `json:"L1L2V"`
    L2L3Voltage SingletonList[PduFloat64] `json:"L2L3V"`
    L3L1Voltage SingletonList[PduFloat64] `json:"L3L1V"`
    L1Current   SingletonList[PduFloat64] `json:"L1I"`
    L2Current   SingletonList[PduFloat64] `json:"L2I"`
    L3Current   SingletonList[PduFloat64] `json:"L3I"`
    NCurrent    SingletonList[PduFloat64] `json:"NI"`
    Frequency   SingletonList[PduFloat64] `json:"freq"`
    PF          SingletonList[PduFloat64] `json:"PF"`
}

func (miv *MainInputValues) GetMetrics() []prom.Collector {
    mainVoltage := prom.NewGaugeVec(
        prom.GaugeOpts{
            Namespace: "stagesmarts",
            Subsystem: "main",
            Name: "voltage",
            Help: "Main input voltage in V",
        },
        []string{"end_a", "end_b"},
    )
    mainVoltage.WithLabelValues("L1", "N").Set(miv.L1Voltage.Value.Value)
    mainVoltage.WithLabelValues("L2", "N").Set(miv.L2Voltage.Value.Value)
    mainVoltage.WithLabelValues("L3", "N").Set(miv.L3Voltage.Value.Value)
    mainVoltage.WithLabelValues("L1", "L2").Set(miv.L1L2Voltage.Value.Value)
    mainVoltage.WithLabelValues("L2", "L3").Set(miv.L2L3Voltage.Value.Value)
    mainVoltage.WithLabelValues("L3", "L1").Set(miv.L3L1Voltage.Value.Value)

    mainCurrent := prom.NewGaugeVec(
        prom.GaugeOpts{
            Namespace: "stagesmarts",
            Subsystem: "main",
            Name: "current",
            Help: "Main input current in A",
        },
        []string{"line"},
    )
    mainCurrent.WithLabelValues("L1").Set(miv.L1Current.Value.Value)
    mainCurrent.WithLabelValues("L2").Set(miv.L2Current.Value.Value)
    mainCurrent.WithLabelValues("L3").Set(miv.L3Current.Value.Value)
    mainCurrent.WithLabelValues("N").Set(miv.NCurrent.Value.Value)

    frequency := prom.NewGaugeVec(
        prom.GaugeOpts{
            Namespace: "stagesmarts",
            Subsystem: "main",
            Name: "frequency",
            Help: "Main input frequency in Hz",
        },
        []string{},
    )
    frequency.WithLabelValues().Set(miv.Frequency.Value.Value)

    return []prom.Collector{
        mainVoltage,
        mainCurrent,
        frequency,
    }
}

type Temperature struct {
    Value PduFloat64 `json:"_"`
}

func (t *Temperature) GetMetrics() []prom.Collector {
    temperature := prom.NewGaugeVec(
        prom.GaugeOpts{
            Namespace: "stagesmarts",
            Subsystem: "system",
            Name: "temparature",
            Help: "System temperature in °C",
        },
        []string{},
    )
    temperature.WithLabelValues().Set(t.Value.Value)

    return []prom.Collector{
        temperature,
    }
}

type Breakers struct {
    MainBreakerStatus  SingletonList[PduBool] `json:"mainBreakerStatus"`
    EStopStatus        SingletonList[PduBool] `json:"eStopStatus"`
    EarthRelayStatus   SingletonList[PduBool] `json:"earthRelayStatus"`
    NeutralRelayStatus SingletonList[PduBool] `json:"neutralRelayStatus"`
}

func (b *Breakers) GetMetrics() []prom.Collector {
    breakerStatus := prom.NewGaugeVec(
        prom.GaugeOpts{
            Namespace: "stagesmarts",
            Subsystem: "main",
            Name: "breaker_status",
            Help: "Breaker status",
        },
        []string{"breaker"},
    )
    breakerStatus.WithLabelValues("main").Set(b.MainBreakerStatus.Value.ToFloat64())
    breakerStatus.WithLabelValues("e-stop").Set(b.EStopStatus.Value.ToFloat64())
    breakerStatus.WithLabelValues("earth-relay").Set(b.EStopStatus.Value.ToFloat64())
    breakerStatus.WithLabelValues("neutral-relay").Set(b.NeutralRelayStatus.Value.ToFloat64())
    return []prom.Collector{
        breakerStatus,
    }
}

type OutputChannels struct {
    Channels []Channel `json:"channel"`
}

func (oc *OutputChannels) GetMetrics(conf *Config) []prom.Collector {
    current := prom.NewGaugeVec(
        prom.GaugeOpts{
            Namespace: "stagesmarts",
            Subsystem: "channels",
            Name: "current",
            Help: "Output channel current",
        },
        []string{"number", "group", "groupNum", "name"},
    )
    inUse := prom.NewGaugeVec(
        prom.GaugeOpts{
            Namespace: "stagesmarts",
            Subsystem: "channels",
            Name: "in_use",
            Help: "Output channel usage indicator",
        },
        []string{"number", "group", "groupNum", "name"},
    )
    for _, c := range oc.Channels {
        name := ""
        for _, cc := range conf.ChannelConfigs {
            if cc.ChannelID == c.ID {
                name = cc.Name
            }
        }
        num := strconv.Itoa(int((c.ID.Group * 6) + c.ID.Channel + 1))
        group := strconv.Itoa(int(c.ID.Group + 1))
        groupNum := strconv.Itoa(int(c.ID.Channel + 1))
        current.WithLabelValues(num, group, groupNum, name).Set(c.Current.Value.Value)
        inUse.WithLabelValues(num, group, groupNum, name).Set(c.InUse.Value.ToFloat64())
    }
    return []prom.Collector{
        current,
        inUse,
    }
}

type Channel struct {
    ID         ChannelID                 `json:"$"`
    Current    SingletonList[PduFloat64] `json:"channelI"`
    MaxCurrent SingletonList[PduFloat64] `json:"channelMaxI"`
    InUse      SingletonList[PduBool]    `json:"channelInUse"`
}

type ChannelID struct {
    Group   uint8 `json:"CIBid,string"`
    Channel uint8 `json:"Channelid,string"`
}

type Config struct {
    Name           string          `json:"name"`
    ChannelConfigs []ChannelConfig `json:"channels"`
}

func (c *Config) GetMetrics() []prom.Collector {
    info := prom.NewGaugeVec(
        prom.GaugeOpts{
            Namespace: "stagesmarts",
            Subsystem: "main",
            Name: "info",
            Help: "System info",
        },
        []string{"name"},
    )
    info.WithLabelValues(c.Name).Set(1)
    return []prom.Collector{
        info,
    }
}

type ChannelConfig struct {
    ChannelID
    Name         string     `json:"name"`
    Supervised   PduBool    `json:"supervised"`
    ThresholdMin PduFloat64 `json:"thresholdmin"`
    ThresholdMax PduFloat64 `json:"thresholdmax"`
}
