package main

func main() {
	rootCmd.Execute()
	sendNotify(notifServerPort)
}
