VER=1.0.1

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