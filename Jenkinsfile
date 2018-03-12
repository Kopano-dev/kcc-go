#!/usr/bin/env groovy

// kcc-go

pipeline {
	agent {
		docker {
			image 'golang:1.9'
			args '-u 0'
		}
	}
	environment {
		GLIDE_VERSION = 'v0.13.1'
		GLIDE_HOME = '/tmp/.glide'
		GOBIN = '/usr/local/bin'
		GOPATH = '/workspace'
		PACKAGE = 'stash.kopano.io/kgol/kcc-go'
	}
	stages {
		stage('Bootstrap') {
			steps {
				sh 'go env'
				sh 'export'
				echo 'Bootstrapping..'
				sh 'mkdir -p \$GOPATH/src/\$PACKAGE && rmdir \$GOPATH/src/\$PACKAGE && ln -sv \$WORKSPACE \$GOPATH/src/\$PACKAGE'
				sh 'curl -sSL https://github.com/Masterminds/glide/releases/download/$GLIDE_VERSION/glide-$GLIDE_VERSION-linux-amd64.tar.gz | tar -vxz -C /usr/local/bin --strip=1'
				sh 'go get -v github.com/golang/lint/golint'
				sh 'go get -v github.com/tebeka/go2xunit'
			}
		}
		stage('Lint') {
			steps {
				echo 'Linting..'
				sh 'cd \$GOPATH/src/\$PACKAGE && golint \$(glide nv) | tee golint.txt || true'
				sh 'cd \$GOPATH/src/\$PACKAGE && go vet \$(glide nv) | tee govet.txt || true'
			}
		}
		stage('Test') {
			steps {
				withCredentials([usernamePassword(credentialsId: 'TEST_CREDENTIALS', usernameVariable: 'TEST_USERNAME', passwordVariable: 'TEST_PASSWORD'), string(credentialsId: 'KOPANO_SERVER_DEFAULT_URI', variable: 'KOPANO_SERVER_DEFAULT_URI')]) {
					echo 'Testing..'
					sh 'echo Kopano Server URI: \$KOPANO_SERVER_DEFAULT_URI'
					sh 'echo Kopano Server Username: \$TEST_USERNAME'
					sh 'cd \$GOPATH/src/\$PACKAGE && glide install'
					sh 'cd \$GOPATH/src/\$PACKAGE && go test -v \$(glide nv) | tee tests.output'
					sh 'cd \$GOPATH/src/\$PACKAGE && go2xunit -fail -input tests.output -output tests.xml'
				}
			}
		}
	}
	post {
		always {
			junit 'tests.xml'
			warnings parserConfigurations: [[parserName: 'Go Lint', pattern: 'golint.txt'], [parserName: 'Go Vet', pattern: 'govet.txt']], unstableTotalAll: '0', messagesPattern: 'don\'t use ALL_CAPS in Go names; use CamelCase'
			cleanWs()
		}
	}
}
