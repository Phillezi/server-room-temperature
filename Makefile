.PHONY: all sensor

all: sensor
	@true

sensor:
	@tinygo build -target pico-w -tags rp2040 -o srt-sensor.uf2 ./cmd/srt-sensor/

