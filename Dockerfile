FROM golang
ADD . /go/src/github.com/cloudendpoints/mixologist

WORKDIR /go/src/github.com/cloudendpoints/mixologist
RUN make clean
RUN make mixologist-bin

ENTRYPOINT ["/go/src/github.com/cloudendpoints/mixologist/mixologist-bin"]
CMD ["-v=1", "-logtostderr=true"]

EXPOSE 9092 
