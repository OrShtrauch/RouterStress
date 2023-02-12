VER=0.7.1
TAG="v$VER"
OWNER="OrShtrauch"
REPO="RouterStress"

dists=("arm" "arm64" "386" "amd64")
os="linux"

wd=$PWD
TMP_DIR="$wd/build/tmp/"
STRESS_DIR="$TMP_DIR/stress"

for arch in "${dists[@]}"; do
    mkdir $TMP_DIR
    mkdir $STRESS_DIR

    cp -r data $STRESS_DIR/
    cp -r containers $STRESS_DIR/
    cp readme.md $STRESS_DIR/
    cp install_docker.sh $STRESS_DIR/

    mkdir $STRESS_DIR/results
    touch $STRESS_DIR/stress.log

    GOOS=$os GOARCH=$arch go build
    cp RouterStress $STRESS_DIR/

    cd $TMP_DIR
    tar -czf RouterStress-$VER-$arch.tar.gz stress/
    mv RouterStress-$VER-$arch.tar.gz "$wd/build/"
    
    cd $wd

    rm -rf $TMP_DIR
done

create_rel_json() {
cat <<EOF
{
    "tag_name":"$TAG",
    "target_commitish":"master",
    "name":"$TAG"
}
EOF
}

git tag -f "$TAG"
# create release from tag
curl \
  -X POST \
  -o /tmp/data.json \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer $TOKEN"\
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "https://api.github.com/repos/$OWNER/$REPO/releases" \
  -d "$(create_rel_json)"

rel_id=$(jq ".id" /tmp/data.json)
rm /tmp/data.json

# upload builds
for arch in "${dists[@]}"; do
    filename="build/RouterStress-$VER-$arch.tar.gz"
    echo $filename
    curl \
    -X POST \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $TOKEN"\
    -H "X-GitHub-Api-Version: 2022-11-28" \
    -H "Content-Type: application/zip" \
    "https://uploads.github.com/repos/$OWNER/$REPO/releases/$rel_id/assets?name=$filename" \
    --data-binary "@$filename"
done