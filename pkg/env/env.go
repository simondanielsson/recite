package env

import "github.com/joho/godotenv"

// Load loads environment variables from files.
func Load(filenames ...string) error {
	return godotenv.Load(filenames...)
}
