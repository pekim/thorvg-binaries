#! /usr/bin/env bash
set -eo pipefail

THORVG_COMMIT=1a43240ec3ffdaa689412e7cd52e83cf8118e2b9
THORVG_DIR=thorvg
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
LIBRARY_DIR=assets
LIBRARY_FILE=$LIBRARY_DIR/libthorvg_${GOOS}_${GOARCH}

# clone if not already cloned
if [ ! -e $THORVG_DIR ]; then
	git clone https://github.com/thorvg/thorvg.git $THORVG_DIR
fi

# ensure desired commit is checked out
pushd $THORVG_DIR
HEAD=$(git rev-parse HEAD)
if [ "$HEAD" != "$THORVG_COMMIT" ]; then
  git pull
  git checkout "${THORVG_COMMIT}"
fi
popd

# build thorvg with C bindings
pushd $THORVG_DIR
meson setup build -Dbindings=capi -Dengines=sw,gl -Dsimd=true
ninja -C build
popd

# copy library and C header
mkdir -p $LIBRARY_DIR
cp $THORVG_DIR/build/src/libthorvg* $LIBRARY_FILE
cp $THORVG_DIR/src/bindings/capi/thorvg_capi.h $LIBRARY_DIR

# strip symbols from the library, reducing the size by about 90%
strip $LIBRARY_FILE
