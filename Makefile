.PHONY: build run test clean docker-build

build:
	go build -o timetable ./cmd/timetable

run: build
	./timetable

test:
	go test ./...

clean:
	rm -f timetable timetable.db

docker-build:
	docker build -t timetable .
