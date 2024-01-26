FROM registry.access.redhat.com/ubi9 AS build

# Config
# ARGs will persist into other images, ENV vars won't. We want these to be available in both.
ENV DIR=/usr/th
ENV BIN_DIR=$DIR/bin
ENV CFG_DIR=$DIR/cfg
ENV SRC_DIR=$DIR/src

RUN dnf -y install go make

RUN mkdir ${DIR} && \
    mkdir ${BIN_DIR} && \
    mkdir ${CFG_DIR} && \
    mkdir ${SRC_DIR}

ADD . ${SRC_DIR}
WORKDIR ${SRC_DIR}

RUN make install-tools

RUN pwd && ls
# build the test binary: outputs to ${SRC_DIR}/vse-sync-testsuite
RUN make build

RUN cp ${SRC_DIR}/vse-sync-testsuite ${BIN_DIR}/collector-tool

RUN dnf remove -y go make && \
	dnf clean all && \
	rm -rf ${TNF_SRC_DIR} && \
	rm -rf ${TEMP_DIR} && \
	rm -rf /root/.cache && \
	rm -rf /root/go/pkg && \
	rm -rf /root/go/src && \
	rm -rf /usr/lib/golang/pkg && \
	rm -rf /usr/lib/golang/src


FROM scratch
COPY --from=build / /
WORKDIR /usr/th/bin
CMD ./collector-tool
