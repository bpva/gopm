TARGET = gopm
SOURCES = $(wildcard cmd/*.go)
BUILD_FLAGS =
build:
	go build $(BUILD_FLAGS) -o $(TARGET) $(SOURCES)
clean:
	rm -f $(TARGET)