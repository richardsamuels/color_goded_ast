.PHONY: clean

all: goout/libcolorgoded.a

goout/libcolorgoded.a: lib.go
	go build -o goout/libcolorgoded.a -buildmode=c-archive lib.go

clean:
	rm -rf goout test
