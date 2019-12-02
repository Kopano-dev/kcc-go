#!/usr/bin/env groovy

// kcc-go

pipeline {
	agent {
		docker {
			image 'golang:1.13'
		}
	}
	environment {
		GOBIN = '/tmp/go-bin'
		GOCACHE = '/tmp/go-build'
	}
	stages {
		stage('Bootstrap') {
			steps {
				echo 'Bootstrapping..'
				sh 'export'
				sh 'go version'
				sh 'go get -v golang.org/x/lint/golint'
				sh 'go get -v github.com/tebeka/go2xunit'
				sh 'go mod vendor'
			}
		}
		stage('Lint') {
			steps {
				echo 'Linting..'
				sh 'PATH=$PATH:$GOBIN golint | tee golint.txt || true'
				sh 'go vet | tee govet.txt || true'
				warnings parserConfigurations: [[parserName: 'Go Lint', pattern: 'golint.txt'], [parserName: 'Go Vet', pattern: 'govet.txt']], unstableTotalAll: '0', messagesPattern: 'don\'t use ALL_CAPS in Go names; use CamelCase'
			}
		}
		stage('Test') {
			steps {
				withCredentials([usernamePassword(credentialsId: 'TEST_CREDENTIALS', usernameVariable: 'TEST_USERNAME', passwordVariable: 'TEST_PASSWORD'), string(credentialsId: 'KOPANO_SERVER_DEFAULT_URI', variable: 'KOPANO_SERVER_DEFAULT_URI')]) {
					echo 'Testing..'
					sh 'echo Kopano Server URI: \$KOPANO_SERVER_DEFAULT_URI'
					sh 'echo Kopano Server Username: \$TEST_USERNAME'
					sh 'go test -v -count=1 | tee tests.output'
					sh 'PATH=$PATH:$GOBIN  go2xunit -fail -input tests.output -output tests.xml'
				}
				junit allowEmptyResults: true, testResults: 'tests.xml'
			}
		}
	}
	post {
		always {
			cleanWs()
		}
	}
}
