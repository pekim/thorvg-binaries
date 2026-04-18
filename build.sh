#! /usr/bin/env bash
set -eo pipefail

THORVG_VERSION=$(<thorvg-version)
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
if [ "$HEAD" != "v$THORVG_VERSION" ]; then
  git fetch
  git checkout "v${THORVG_VERSION}"
fi
popd

# build thorvg with C bindings
pushd $THORVG_DIR
meson setup build -Dbindings=capi -Dengines=cpu,gl -Dsimd=true
ninja -C build
popd

# copy library and C header
mkdir -p $LIBRARY_DIR
if [ -f $THORVG_DIR/build/src/libthorvg-1.so.$THORVG_VERSION ]; then
  cp $THORVG_DIR/build/src/libthorvg-1.so.$THORVG_VERSION $LIBRARY_FILE
  # strip symbols from the library, reducing the size by about 90%
  strip $LIBRARY_FILE
fi
if [ -f $THORVG_DIR/build/src/libthorvg-1.1.dylib ]; then
  cp $THORVG_DIR/build/src/libthorvg-1.1.dylib $LIBRARY_FILE
fi
cp $THORVG_DIR/src/bindings/capi/thorvg_capi.h $LIBRARY_DIR

# list outputs
ls -l $LIBRARY_DIR
