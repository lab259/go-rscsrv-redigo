
GOPATH=$(CURDIR)/.gopath
GOPATHCMD=GOPATH=$(GOPATH)

COVERDIR=$(CURDIR)/.cover
COVERAGEFILE=$(COVERDIR)/cover.out

.PHONY: deps deps-ci coverage coverage-ci test test-watch coverage coverage-html

test:
	@${GOPATHCMD} ginkgo --failFast ./...

test-watch:
	@${GOPATHCMD} ginkgo watch -cover -r ./...

coverage-ci:
	@mkdir -p $(COVERDIR)
	@${GOPATHCMD} ginkgo -r -covermode=count --cover --trace ./
	@echo "mode: count" > "${COVERAGEFILE}"
	@find . -type f -name *.coverprofile -exec grep -h -v "^mode:" {} >> "${COVERAGEFILE}" \; -exec rm -f {} \;

coverage: coverage-ci
	@sed -i -e "s|_$(CURDIR)/|./|g" "${COVERAGEFILE}"

coverage-html:
	@$(GOPATHCMD) go tool cover -html="${COVERAGEFILE}" -o .cover/report.html

deps:
	@mkdir -p ${GOPATH}
	@go list -f '{{join .Deps "\n"}}' . | xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' | GOPATH=${GOPATH} xargs go get -v -t
	@go list -f '{{join .TestImports "\n"}}' . | xargs go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' | GOPATH=${GOPATH} xargs go get -v -t

deps-ci:
	-go get -v -t ./...
