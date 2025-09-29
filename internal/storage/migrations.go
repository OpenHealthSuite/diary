package storage

import (
	"embed"
	"regexp"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:embed sqlc/schema/*.sql
var sqlFiles embed.FS

func getPgMigrations() [][]string {
	migrationfiles, err := sqlFiles.ReadDir("sqlc/schema")
	if err != nil {
		log.Fatal().Err(err).Msg("Fatal error reading migrations dir")
	}
	data := [][]string{}
	re := regexp.MustCompile(`^\d{5}$`)
	for _, fl := range migrationfiles {
		if !fl.IsDir() {
			fnm_prts := strings.Split(fl.Name(), "_")
			key := fnm_prts[0]
			if !re.MatchString(key) {
				log.Warn().Msgf("Malformed filename in migrations: %s", fl.Name())
				continue
			}
			migration, err := sqlFiles.ReadFile("sqlc/schema/" + fl.Name())
			if err != nil {
				log.Fatal().Err(err).Msgf("Fatal error reading file %s", fl.Name())
			}
			data = append(data, []string{key, string(migration)})
		}
	}
	slices.SortFunc(data, func(a []string, b []string) int {
		return strings.Compare(a[0], b[0])
	})
	return data
}
