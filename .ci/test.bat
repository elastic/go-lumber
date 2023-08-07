mkdir -p build
SET OUT_FILE=build\output-report.out
go test "./..." -v > %OUT_FILE% | type %OUT_FILE%

go install github.com/jstemmer/go-junit-report/v2@latest
go-junit-report > build\junit-%RUNNER_OS%.xml < %OUT_FILE%
