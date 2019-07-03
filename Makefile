COVERDIR=$(CURDIR)/.cover
COVERAGEFILE=$(COVERDIR)/cover.out

dcup:
	@docker-compose up -d

test: dcup
	@ginkgo --failFast ./...

test-watch: dcup
	@ginkgo watch -cover -r ./...

coverage-ci: dcup
	@mkdir -p $(COVERDIR)
	@ginkgo -r -covermode=count --cover --trace ./
	@echo "mode: count" > "${COVERAGEFILE}"
	@find . -type f -name *.coverprofile -exec grep -h -v "^mode:" {} >> "${COVERAGEFILE}" \; -exec rm -f {} \;

coverage: coverage-ci
	@sed -i -e "s|_$(CURDIR)/|./|g" "${COVERAGEFILE}"

coverage-html:
	@go tool cover -html="${COVERAGEFILE}" -o .cover/report.html


.PHONY: dcup test test-watch coverage coverage-ci coverage-html