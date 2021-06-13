package args

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

func TestParseArgsWithOptionalCmd(t *testing.T) {
  stringArgs := []string{
      "service=shell",
      "cmd*",
  }
  
  args := ParseAuthorArgs(stringArgs)

  if assert.NotNil(t, args) {
      if assert.NotNil(t, args.Service) {
        assert.Equal(t, "shell", *args.Service)
      }
      if assert.NotNil(t, args.Cmd) {
        assert.Equal(t, "", *args.Cmd) // Should be set, but empty
      }
      assert.Equal(t, []string(nil), args.CmdArg) // Should be an empty slice
  }
}


func TestParseArgsWithShowRunningConfig(t *testing.T) {
  stringArgs := []string{
      "service=shell",
      "cmd=show",
      "cmd-arg=running-config",
      "cmd-arg=<cr>",
  }
  
  args := ParseAuthorArgs(stringArgs)

  if assert.NotNil(t, args) {
      if assert.NotNil(t, args.Service) {
        assert.Equal(t, "shell", *args.Service)
      }
      if assert.NotNil(t, args.Cmd) {
        assert.Equal(t, "show", *args.Cmd)
      }
      assert.Equal(t, "show running-config <cr>", args.AsShellCommand())
  }
}


func TestParseArgsWithClearInterfaceCounters(t *testing.T) {
  stringArgs := []string{
      "service=shell",
      "cmd=clear",
      "cmd-arg=interface",
      "cmd-arg=counters",
      "cmd-arg=GigabitEthernet",
      "cmd-arg=1/12",
      "cmd-arg=<cr>",
  }
  
  args := ParseAuthorArgs(stringArgs)

  if assert.NotNil(t, args) {
      if assert.NotNil(t, args.Service) {
        assert.Equal(t, "shell", *args.Service)
      }
      if assert.NotNil(t, args.Cmd) {
        assert.Equal(t, "clear", *args.Cmd)
      }
      assert.Equal(t, "clear interface counters GigabitEthernet 1/12 <cr>", args.AsShellCommand())
  }
}
