package servers

func StartServers() {
	go InitPostgresServer()
	go InitMongoServer()
	go InitAuthServer()

	select {}
}
