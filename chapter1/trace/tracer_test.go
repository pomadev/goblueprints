package trace

import (
	"bytes"
	"testing"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	tracer := New(&buf)
	tracer.Trace("こんにちは、traceパッケージ")
	if buf.String() != "こんにちは、traceパッケージ\n" {
		t.Errorf("'%s'という誤った文字列が出力されました。", buf.String())
	}

}

func TestNil(t *testing.T) {
	var tracer Tracer
	tracer.Trace("データ")
}
