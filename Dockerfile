FROM golang
ADD . /go/src/somnacin-internal/mixologist

WORKDIR /go/src/somnacin-internal/mixologist
RUN make clean
RUN make mixologist-bin

ENTRYPOINT ["/go/src/somnacin-internal/mixologist/mixologist-bin"]
CMD ["-v=1", "-logtostderr=true"]

EXPOSE 9092 
