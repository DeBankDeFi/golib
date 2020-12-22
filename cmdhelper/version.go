package cmdhelper

import (
	"github.com/DeBankDeFi/golib/shared"

	"github.com/XSAM/go-hybrid/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Version() *cobra.Command {
	cmd := cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			log.BgLogger().Info("app info", zap.Any("info", shared.AppInfo()))
		},
	}
	return &cmd
}