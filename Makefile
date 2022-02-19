SHELL=cmd
STRIPE_SECRET=sk_test_51KEmG3JxV5c4eSlbMR273GPRyB1xEY2ml88GhgXsnJuwyuIJND8VBuLqLYRuE5jV9q3Bud14Y4hxtCaokA0xsYxz00sXwhL7aY
STRIPE_KEY=pk_test_51KEmG3JxV5c4eSlbjMSzW69ht79i17P8luMgFxl8fyX3uFUKqeckTqxJTTLtgpktBGBCg3JYEnDzS0YisvMj1REX00J7vdvjyB
GOSTRIPE_PORT=4000
API_PORT=4001
DSN="leetrent:CaseyPo0h@tcp(localhost:3306)/widgets?parseTime=true&tls=false"
SMTP_HOST=smtp.mailtrap.io
SMTP_PORT=587
SMTP_USERNAME=e1717335a62b27
SMTP_PASSWORD=ea8e6aa5148a8d
SECRET_KEY=TPV7G4yOjIIJvr1plajBkRvlRI0HUUVU
FRONTEND_URL=http://localhost:4000

## build: builds all binaries
build: clean build_front build_back
	@echo All binaries built!

## clean: cleans all binaries and runs go clean
clean:
	@echo Cleaning...
	@echo y | DEL /S dist
	@go clean
	@echo Cleaned and deleted binaries

## build_front: builds the front end
build_front:
	@echo Building front end...
	@go build -o dist/gostripe.exe ./cmd/web
	@echo Front end built!

## build_invoice: builds the invoice microservice
build_invoice:
	@echo Building invoice microservice...
	@go build -o dist/invoice.exe ./cmd/micro/invoice
	@echo Invoice microservice built!

## build_back: builds the back end
build_back:
	@echo Building back end...
	@go build -o dist/gostripe_api.exe ./cmd/api
	@echo Back end built!

## start: starts front and back end
start: start_front start_back start_invoice

## start_front: starts the front end
start_front: build_front
	@echo Starting the front end...
	set STRIPE_KEY=${STRIPE_KEY}&& set STRIPE_SECRET=${STRIPE_SECRET}&& set SECRET_KEY=${SECRET_KEY}&& set FRONTEND_URL=${FRONTEND_URL}&& start /B .\dist\gostripe.exe -dsn=${DSN}
	@echo Front end running!

## start_back: starts the back end
start_back: build_back
	@echo Starting the back end...
	set STRIPE_KEY=${STRIPE_KEY}&& set STRIPE_SECRET=${STRIPE_SECRET}&& set SMTP_HOST=${SMTP_HOST}&& set SMTP_PORT=${SMTP_PORT}&& set SMTP_USERNAME=${SMTP_USERNAME}&& set SMTP_PASSWORD=${SMTP_PASSWORD}&& set SECRET_KEY=${SECRET_KEY}&& set FRONTEND_URL=${FRONTEND_URL}&& start /B .\dist\gostripe_api.exe -dsn=${DSN}
	@echo Back end running!

## start_invoice: starts the invoice microservice
start_invoice: build_invoice
	@echo Starting the invoice microservice...
	set SMTP_HOST=${SMTP_HOST}&& set SMTP_PORT=${SMTP_PORT}&& set SMTP_USERNAME=${SMTP_USERNAME}&& set SMTP_PASSWORD=${SMTP_PASSWORD}&& set FRONTEND_URL=${FRONTEND_URL}&& start /B .\dist\invoice.exe
	@echo Invoice microservice running!

## stop: stops the front and back end
stop: stop_front stop_back stop_invoice
	@echo All applications stopped

## stop_front: stops the front end
stop_front:
	@echo Stopping the front end...
	@taskkill /IM gostripe.exe /F
	@echo Stopped front end

## stop_back: stops the back end
stop_back:
	@echo Stopping the back end...
	@taskkill /IM gostripe_api.exe /F
	@echo Stopped back end

## stop_invoice: stops the invoice microservice
stop_invoice:
	@echo Stopping the invoice microservice...
	@taskkill /IM invoice.exe /F
	@echo Stopped invoice microservice