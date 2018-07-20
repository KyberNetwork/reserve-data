#!/bin/bash

set -euo pipefail
readonly mode="${1:-}"
readonly github_url="https://github.com/KyberNetwork/reserve-data/pull/"

print_usage() {
    cat <<EOF
Usage:
  ./pr_diff commit <commit-id-1> <commit-id-2>
  ./pr_diff timestamp <timestamp-1> <timestamp-2>
EOF
}

if [[ -z "$mode" ]]; then
    print_usage
    exit 1
fi

case $mode in
    "timestamp")
	if [[ $# -ne 3 ]]; then
	    print_usage
	    exit 1
	fi
	after=$2
	before=$3
	pr_commit_ids=($(git log --oneline --after="$after" --before="$before" \
			     | awk '/Merge pull request/ {print $1}'))
	;;
    "commit")
	if [[ $# -ne 3 ]]; then
	    print_usage
	    exit 1
	fi
	first_commit=$2
	last_commit=$3
	pr_commit_ids=($(git log --oneline "$first_commit".."$last_commit" \
			     | awk '/Merge pull request/ {print $1}'))
	;;
    "*")
	print_usage
	exit 1
esac

for commit_id in "${pr_commit_ids[@]}"; do
    git show --format="%B" "$commit_id" |\
	awk -v github_url="$github_url" 'NR == 1 {$4="["$4"]("github_url substr($4,2)")"; print "* "$0; next}; /^diff --/{exit}; {print "  "$0}'
done
