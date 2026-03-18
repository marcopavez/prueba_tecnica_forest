package config

import "os"

func GetJWTSecret() string {

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-for-testing"
	}

	return jwtSecret
}

func GetAdminCredentials() string {

	adminCredentials := os.Getenv("ADMIN_CREDENTIALS")
	if adminCredentials == "" {
		// default: admin:password
		adminCredentials = "YWRtaW46cGFzc3dvcmQ="
	}

	return adminCredentials
}

func GetDbPath() string {

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./bike_rental.db"
	}

	return dbPath
}

func GetPort() string {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return port
}
