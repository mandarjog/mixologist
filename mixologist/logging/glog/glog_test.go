package glog

import (
	"bytes"
	"flag"
	"github.com/golang/glog"
	"io"
	"os"
	"somnacin-internal/mixologist/mixologist"
	"strings"
	"testing"
)

func TestLog(t *testing.T) {
	tests := []struct {
		name string
		in   mixologist.LogEntry
		want string
	}{
		{
			name: "Log Entry with List",
			in: mixologist.LogEntry{
				Name:          "test-log",
				Resource:      mixologist.Resource{Type: "api"},
				Labels:        map[string]string{"test-label": "GET"},
				Severity:      "INFO",
				StructPayload: map[string]interface{}{"list": []interface{}{9453, "test"}, "struct": map[string]interface{}{"field": 9e3}},
			},
			want: "{\"logName\":\"test-log\",\"timestamp\":\"0001-01-01T00:00:00Z\",\"id\":\"\",\"resource\":{\"type\":\"api\"},\"labels\":{\"test-label\":\"GET\"},\"severity\":\"INFO\",\"structPayload\":{\"list\":[9453,\"test\"],\"struct\":{\"field\":9000}}}",
		},
	}

	// we want to ensure glog writes to stderr so we can redirect for
	// test validation of log generation
	flag.Set("logtostderr", "true")

	for _, v := range tests {
		l := (&builder{}).Build(mixologist.Config{})

		sl := glog.Stats.Info.Lines() // validate lines

		old := os.Stderr // for restore
		r, w, _ := os.Pipe()
		os.Stderr = w // redirecting

		// copy over the output from stderr
		outC := make(chan string)
		go func() {
			var buf bytes.Buffer
			io.Copy(&buf, r)
			outC <- buf.String()
		}()

		l.Log(v.in)
		l.Flush()

		// back to normal state
		w.Close()
		os.Stderr = old
		got := <-outC

		if gotLines := glog.Stats.Info.Lines() - sl; gotLines != 1 {
			t.Errorf("%s: got %v lines, want 1", v.name, gotLines)
		}

		if trim(got) != v.want {
			t.Errorf("%s: got %s, want %s", v.name, trim(got), v.want)
		}
	}
}

// trims glog line prefix stuff off of output log lines (glog prefix ends with ']')
func trim(s string) string {
	return strings.TrimRight(strings.TrimLeft(strings.SplitN(s, "]", 2)[1], " "), "\n")
}
