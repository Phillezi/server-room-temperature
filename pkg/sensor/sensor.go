package sensor

type Sensor interface {
	Read() (int32, error)
}
