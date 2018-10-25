#!/bin/bash
set -e

check_prerequisites() {
	[[ "$GOPATH" == "" ]] && \
		errors=(${errors[@]} "GOPATH env missing")
	
	[[ -x "$GOPATH/bin/dep" ]] || \
		errors=("${errors[@]}" "dep not found in \"$GOPATH/bin/\"")

	if [[ "${#errors[@]}" > 0 ]]; then
		echo "Errors:"
		for error in "${errors[@]}"; do
			echo "  $error"
		done
		return 1
	fi
	return 0
}

check_versions() {
	VERS="${LATEST_RELEASE}\n${MAIN_VER}"
	DKR_TAG="master"

	# For tagged commits
	if [ "$(git describe --tags)" = "$(git describe --tags --abbrev=0)" ] ; then
		if [ $(printf ${VERS} | uniq | wc -l) -gt 1 ] ; then
			echo "This is a release, all versions should match"
			return 1
		fi
		DKR_TAG="latest"
	else
		if [ $(printf ${VERS} | uniq | wc -l) -eq 1 ] ; then
			echo "Please increment the version in main.go"
			return 1
		fi
		if [ "$(printf ${VERS} | sort -V | tail -n 1)" != "${MAIN_VER}" ] ; then
			echo "Please increment the version in main.go"
			return 1
		fi
	fi
}

LATEST_RELEASE=$(git describe --tags --abbrev=0 | sed "s/^v//g")
MAIN_VER=$(grep "\t*version *= " main.go | sed 's/\t*version *= //g' | sed 's/"//g')

check_prerequisites || exit 1
check_versions || exit 1

echo "Installing Dependencies..."
$GOPATH/bin/dep ensure

echo "Linting..."
gometalinter --vendor ./...


echo "Building..."
mkdir bin 2>/dev/null || true
go build -o bin/docker-zfs-plugin .
