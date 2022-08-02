FROM golang:1.18 as build

WORKDIR /build

RUN go env -w GOPROXY=direct
ADD go.mod go.sum ./
RUN go mod download

ADD . .
RUN go build -o ./main

FROM public.ecr.aws/lambda/go:1

COPY --from=build /build/main ${LAMBDA_TASK_ROOT}

# Set the CMD to your handler (could also be done as a parameter override outside of the Dockerfile)
CMD [ "main" ]
