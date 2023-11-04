package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/khulnasoft-lab/codeai-backend/application/config"
	"github.com/khulnasoft-lab/codeai-backend/internal/testutil"
)

func Test_shouldSetLogLevelViaFlag(t *testing.T) {
	args := []string{"codeai-backend", "-l", "debug"}
	_, _ = parseFlags(args, config.New())
	assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())
}

func Test_shouldSetLogFileViaFlag(t *testing.T) {
	args := []string{"codeai-backend", "-f", "a.txt"}
	t.Cleanup(func() {
		config.CurrentConfig().DisableLoggingToFile()

		err := os.Remove("a.txt")
		if err != nil {
			t.Logf("Error when trying to cleanup logfile: %e", err)
		}
	})

	_, _ = parseFlags(args, config.New())
	assert.Equal(t, config.CurrentConfig().LogPath(), "a.txt")
}

func Test_shouldSetOutputFormatViaFlag(t *testing.T) {
	args := []string{"codeai-backend", "-o", config.FormatHtml}
	_, _ = parseFlags(args, config.New())
	assert.Equal(t, config.FormatHtml, config.CurrentConfig().Format())
}

func Test_shouldShowUsageOnUnknownFlag(t *testing.T) {
	args := []string{"codeai-backend", "-unknown", config.FormatHtml}

	output, err := parseFlags(args, config.New())

	assert.True(t, strings.Contains(output, "Usage of codeai-backend"))
	assert.NotNil(t, err)
}

func Test_shouldDisplayLicenseInformationWithFlag(t *testing.T) {
	args := []string{"codeai-backend", "-licenses"}
	output, _ := parseFlags(args, config.New())
	assert.True(t, strings.Contains(output, "License information"))
}

func Test_shouldReturnErrorWithVersionStringOnFlag(t *testing.T) {
	args := []string{"codeai-backend", "-v"}
	output, err := parseFlags(args, config.New())
	assert.Error(t, err)
	assert.Empty(t, output)
	assert.Equal(t, config.Version, err.Error())
}

func Test_shouldSetLoadConfigFromFlag(t *testing.T) {
	file, err := os.CreateTemp(".", "configFlagTest")
	if err != nil {
		assert.Fail(t, "Couldn't create test file")
	}
	defer func(file *os.File) {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}(file)

	_, err = file.Write([]byte("AA=Bb"))
	if err != nil {
		assert.Fail(t, "Couldn't write to test file")
	}
	args := []string{"codeai-backend", "-c", file.Name()}

	t.Setenv("Bb", "")

	_, _ = parseFlags(args, config.New())
	assert.Equal(t, "Bb", os.Getenv("AA"))
}

func Test_shouldSetReportErrorsViaFlag(t *testing.T) {
	testutil.UnitTest(t)
	args := []string{"codeai-backend"}
	_, _ = parseFlags(args, config.New())

	assert.False(t, config.CurrentConfig().IsErrorReportingEnabled())

	args = []string{"codeai-backend", "-reportErrors"}
	_, _ = parseFlags(args, config.New())
	assert.True(t, config.CurrentConfig().IsErrorReportingEnabled())
}

func Test_ConfigureLoggingShouldAddFileLogger(t *testing.T) {
	testutil.UnitTest(t)
	logPath := t.TempDir()
	logFile := filepath.Join(logPath, "a.txt")
	config.CurrentConfig().SetLogPath(logFile)
	t.Cleanup(func() {
		config.CurrentConfig().DisableLoggingToFile()
	})

	config.CurrentConfig().ConfigureLogging(nil)
	log.Error().Msg("test")

	assert.Eventuallyf(t, func() bool {
		bytes, err := os.ReadFile(config.CurrentConfig().LogPath())
		fmt.Println("Read file " + config.CurrentConfig().LogPath())
		if err != nil {
			return false
		}
		fmt.Println("Read bytes:" + string(bytes)) // no logger usage here
		return len(bytes) > 0
	}, 2*time.Second, 10*time.Millisecond, "didn't write to logfile")
}
