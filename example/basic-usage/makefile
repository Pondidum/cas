


dist/index.js: $(shell find src/ -type f | ../cas fetch)
	@echo "Transpiling TypeScript"
	@sleep 5s

	@mkdir -p dist
	@echo "compiled at $(shell date)" > "$@"
	@../../cas artifact "$<" "$@"