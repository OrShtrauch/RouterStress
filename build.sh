VER=0.8.1
TAG="v$VER"
OWNER="OrShtrauch"
REPO="RouterStress"

dists=("arm" "arm64" "386" "amd64")
os="linux"

wd=$PWD
TMP_DIR="$wd/build/tmp/"
STRESS_DIR="$TMP_DIR/stress"
RESULTS_DIR="$STRESS_DIR/results"

tmpfile=/tmp/data.json

build() {
    # building for each architecture
    # and creating tar archive with the required folder structure
    for arch in "${dists[@]}"; do        
        mkdir -p $RESULTS_DIR

        cp -r data $STRESS_DIR/
        cp -r containers $STRESS_DIR/
        cp readme.md $STRESS_DIR/
        cp install_docker.sh $STRESS_DIR/

        GOOS=$os GOARCH=$arch go build
        cp RouterStress $STRESS_DIR/

        cd $TMP_DIR
        tar -czf RouterStress-$VER-$arch.tar.gz stress/
        mv RouterStress-$VER-$arch.tar.gz "$wd/build/"

        cd $wd

        #rm -rf "$wd/build/*"
        rm "$wd/RouterStress"
    done

}

create_rel_json() {
cat <<EOF
{
    "tag_name":"$TAG",
    "target_commitish":"master",
    "name":"$TAG"
}
EOF
}

create_release() {
    curl \
    -X POST \
    -o /tmp/data.json \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $GH_TOKEN"\
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "https://api.github.com/repos/$OWNER/$REPO/releases" \
    -d "$(create_rel_json)"

    rel_id=$(jq ".id" $tmpfile)
    rm $tmpfile
}


upload_builds() {
    for arch in "${dists[@]}"; do
        filename="RouterStress-$VER-$arch.tar.gz"
        filepath="build/$filename"
        echo $filename
        curl \
        -X POST \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer $GH_TOKEN"\
        -H "X-GitHub-Api-Version: 2022-11-28" \
        -H "Content-Type: application/zip" \
        "https://uploads.github.com/repos/$OWNER/$REPO/releases/$rel_id/assets?name=$filename" \
        --data-binary "@$filepath"
    done
}

delete_old_release() {
    curl \
    -o /tmp/data.json \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $GH_TOKEN" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "https://api.github.com/repos/$OWNER/$REPO/releases"

    rel_ids=$(jq ".[].id" $tmpfile)

    # iterating over all releases to find the one with the same name
    for id in ${rel_ids[@]}
    do
        old_id=$(jq -r --arg NAME "$TAG"  '.[] | select(.name==$NAME) | .id' $tmpfile)

        if [ ! -z $old_id ]; then 
            curl \
            -X DELETE \
            -H "Accept: application/vnd.github+json" \
            -H "Authorization: Bearer $GH_TOKEN" \
            -H "X-GitHub-Api-Version: 2022-11-28" \
            https://api.github.com/repos/$OWNER/$REPO/releases/$old_id
        fi
    done

    rm $tmpfile
}

run() {
    rm $tmpfile
    # deleting any other release by the same name
    delete_old_release
    # updating the tag
    git tag -f "$TAG"
    # creating a new release in github
    create_release
    # buidlig and creating tar archives
    build
    # upload build to release as assests
    upload_builds
    # cleanup build folder
    rm -r $wd/build/*
}

run
