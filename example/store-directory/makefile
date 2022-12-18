

dist/.archive: $(shell find src/ -type f | ../../cas fetch --verbose)
	@echo "Transpiling TypeScript"
	@sleep 5s

	@mkdir -p dist
	@echo "compiled at $(shell date)" > "dist/index.js"
	@echo "types compiled at $(shell date)" > "dist/index.d.ts"
	@echo "sourcemaps compiled at $(shell date)" > "dist/index.map.js"

	@../../cas artifact "$<" "$@"
