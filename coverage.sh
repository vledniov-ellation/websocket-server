#!/bin/sh
# Generate test coverage statistics for Go packages.
#
help_msg=" Usage: coverage.sh\n
     --html      Additionally create HTML report and open it in browser\n
     --intgr     Run only tests in test files with build tag intgr or without tags.\n
     --all       Run all tests.\n
     --help      Show this message and exit\n
"

workdir=.cover
profile="$workdir/cover.out"
mode=count
show_cover_report_html=0
tag=""
test_failed=0

generate_cover_data() {
    rm -rf "$workdir"
    mkdir "$workdir"

    if ! go test -cover -covermode="$mode" -coverprofile="$profile" -coverpkg=./... ./... -tags="$tag"
    then
        # mark build as failed
        test_failed=1
    fi
}

show_cover_report() {
    go tool cover -${1}="$profile"
}

parse_cmd_flags() {
    for i in "$@"
    do
        case "$i" in
            "")
            ;;
            --html)
                show_cover_report_html=1
            ;;
            --intgr)
                tag="$tag intgr"
            ;;
            --all)
                tag="all intgr"
            ;;
            --help)
                echo -e $help_msg
                exit 0
                ;;
            *)
            echo >&2 "error: invalid option: $1"; exit 1 ;;
        esac
    done
}

parse_cmd_flags $@
generate_cover_data
show_cover_report func

if [ "$show_cover_report_html" -eq 1 ]; then
    show_cover_report html
fi

exit $test_failed
