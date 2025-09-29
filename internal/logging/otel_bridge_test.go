package logging

import (
    "testing"
    "time"

    "github.com/sirupsen/logrus"
    sdklog "go.opentelemetry.io/otel/sdk/log"
)

func TestOtelHookLevels(t *testing.T) {
    lp := sdklog.NewLoggerProvider()
    hook := NewOtelHook(lp)
    lvls := hook.Levels()
    if len(lvls) == 0 { t.Fatal("levels") }
}

func TestAddOtelHook_NoPanic(t *testing.T) {
    lp := sdklog.NewLoggerProvider()
    logger := logrus.New()
    AddOtelHook(logger, lp)
    logger.Info("x")
}

func TestAddOtelHook_WithNilProvider(t *testing.T) {
    logger := logrus.New()
    AddOtelHook(logger, nil)
    // Should not panic and hook should be added
    logger.Info("test with nil provider")
}

func TestOtelHook_Fire(t *testing.T) {
    hook := NewOtelHook(nil)
    entry := &logrus.Entry{
        Time:    time.Now(),
        Level:   logrus.InfoLevel,
        Message: "test message",
        Data: logrus.Fields{
            "key": "value",
            "trace_id": "test-trace-id",
            "span_id": "test-span-id",
        },
    }
    
    // Should not panic with nil logger
    err := hook.Fire(entry)
    if err != nil {
        t.Errorf("expected no error, got: %v", err)
    }
}

func TestConvertLevel(t *testing.T) {
    hook := NewOtelHook(nil)
    
    tests := []logrus.Level{
        logrus.ErrorLevel,
        logrus.WarnLevel,
        logrus.InfoLevel,
        logrus.DebugLevel,
        logrus.TraceLevel,
    }
    
    for _, level := range tests {
        severity := hook.convertLevel(level)
        // Just verify it doesn't panic and returns something
        if severity == 0 {
            t.Errorf("convertLevel(%v) returned zero severity", level)
        }
    }
}

func TestToString_IndirectlyThroughFire(t *testing.T) {
    // This tests the toString function indirectly through Fire
    hook := NewOtelHook(nil)
    entry := &logrus.Entry{
        Time:    time.Now(),
        Level:   logrus.InfoLevel,
        Message: "test",
        Data: logrus.Fields{
            "string": "value",
            "int":    42,
            "bool":   true,
            "slice":  []string{"a", "b"},
        },
    }
    
    err := hook.Fire(entry)
    if err != nil {
        t.Errorf("expected no error, got: %v", err)
    }
}


