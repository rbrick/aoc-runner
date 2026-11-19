package cmd

import (
	"github.com/rbrick/aoc-runner/internal/database"
	"github.com/rbrick/aoc-runner/internal/session"
	"github.com/rbrick/aoc-runner/internal/web"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Starts the HTTP RESTful API for aoc-runner",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		db, err := gorm.Open(sqlite.Open(viper.GetString("server.database")))

		cobra.CheckErr(err)

		// migrate database
		cobra.CheckErr(database.Migrate(ctx, db))
		// seed database
		cobra.CheckErr(database.Seed(ctx, db))

		// cookies are read from the browser when /auth/login is called
		aocSession := session.NewSession(session.ReadBrowserCookies)

		cobra.CheckErr(web.Start(ctx, db, aocSession, viper.GetString("server.host")))
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)
}
