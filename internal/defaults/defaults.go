package defaults

const (
	DefaultNATSURL            = "nats://localhost:4222"
	DefaultNATSHost           = "localhost:4222"
	DefaultNATSWriterUser     = "temperature-writer"
	DefaultNATSWriterPassword = "CHANGE-ME-WRITER"
	DefaultNATSDashboardUser  = "temperature-dashboard"
	DefaultNATSDashboardPassword = "CHANGE-ME-DASHBOARD"
	DefaultNATSReaderUser     = "temperature-reader"
	DefaultNATSReaderPassword = "CHANGE-ME-READER"
	DefaultNATSWSURL          = "ws://localhost:9222"
	DefaultNATSTopic          = "serverroom.temperature.room1.sensor1"
	DefaultNATSStream         = "SERVER_ROOM_TEMPERATURE"
	DefaultHTTPAddr           = ":8080"
	DefaultSerialPort         = "/dev/ttyACM0"
)
