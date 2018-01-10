PREFIX=/home/${USER}/COLAS/src/ZMQ/zmqlibs
echo $PREFIX
rm -rf ${PREFIX}/*
./configure --prefix=${PREFIX}
make -j 4
make check
make install 
