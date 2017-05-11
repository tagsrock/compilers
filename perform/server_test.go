package perform

import (
	"testing"

	"time"

	"github.com/monax/compilers/definitions"
	"github.com/stretchr/testify/assert"
)

func TestStartServer(t *testing.T) {
	// Check server starting is (probably) idempotent
	for i := 0; i < 2; i++ {
		closer, ch := StartServer(":9099", "", "", "")
		// It pains me to do this, but doing it via a ready channel turns out to be
		// a huge yak shave
		time.Sleep(time.Second)
		err := closer.Close()
		assert.NoError(t, err)
		assertShutdown(t, ch)
	}
}

// This is very crude smoke test but it's better than nothing
func TestRoutesRunning(t *testing.T) {
	compiler := definitions.Compiler{
		Lang:   definitions.SOLIDITY,
		Config: definitions.LangConfig{},
	}

	// Try compiler root route
	closer, ch := StartServer(":9099", "", "", "")
	_, err := requestResponse(compiler.CompilerRequest("", nil, "",
		true, nil), "http://:9099")
	assert.NoError(t, err)

	// Try binaries route
	_, err = requestBinaryResponse(&definitions.BinaryRequest{}, "http://:9099/binaries")
	assert.NoError(t, err)
	err = closer.Close()
	assert.NoError(t, err)
	assertShutdown(t, ch)
}

func assertShutdown(t *testing.T, ch chan error) {
	err := <-ch
	assert.Contains(t, err.Error(), "use of closed network connection")
}
