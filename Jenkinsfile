pipeline {
    agent any

    options {
        disableConcurrentBuilds()
        timestamps()
        skipDefaultCheckout(false)
    }

    environment {
        BUILD_SERVER = '192.168.18.8'
        BUILD_USER = 'root'
        REMOTE_ROOT = '/opt/jenkins'
    }

    stages {
        stage('Add Host Key') {
            steps {
                sh '''
                    set -e
                    ssh-keyscan -H "${BUILD_SERVER}" >> ~/.ssh/known_hosts
                '''
            }
        }
        stage('Prepare remote workspace') {
            steps {
                sh '''
                    set -e
                    REMOTE_DIR="${REMOTE_ROOT}/dawg-${BUILD_TAG}"
                    ssh -o BatchMode=yes "${BUILD_USER}@${BUILD_SERVER}" \
                        "mkdir -p '${REMOTE_DIR}'"
                    tar \
                        --exclude='.git' \
                        --exclude='node_modules' \
                        --exclude='engine/node_modules' \
                        --exclude='desktop/dist' \
                        --exclude='desktop/src-tauri/target' \
                        --exclude='.cache' \
                        -czf - . |
                        ssh -o BatchMode=yes "${BUILD_USER}@${BUILD_SERVER}" \
                            "tar -xzf - -C '${REMOTE_DIR}'"
                '''
            }
        }

        stage('Remote validation') {
            steps {
                sh '''
                    set -e
                    REMOTE_DIR="${REMOTE_ROOT}/dawg-${BUILD_TAG}"
                    ssh -o BatchMode=yes "${BUILD_USER}@${BUILD_SERVER}" \
                        "cd '${REMOTE_DIR}/engine' && \
                         go version && \
                         node --version && \
                         npm --version && \
                         rustc --version && \
                         cargo --version && \
                         npm ci && \
                         npx --no-install playwright install chromium && \
                         test -z \"\\$(gofmt -l .)\" && \
                         go vet ./... && \
                         go test ./... && \
                         go build ./cmd/dawg"
                    ssh -o BatchMode=yes "${BUILD_USER}@${BUILD_SERVER}" \
                        "cd '${REMOTE_DIR}/desktop' && \
                         npm ci && \
                         npx --no-install biome check . && \
                         npm run build"
                    ssh -o BatchMode=yes "${BUILD_USER}@${BUILD_SERVER}" \
                        "cd '${REMOTE_DIR}/desktop/src-tauri' && \
                         cargo check --locked"
                '''
            }
        }

        stage('Collect build artifacts') {
            steps {
                sh '''
                    set -e
                    REMOTE_DIR="${REMOTE_ROOT}/dawg-${BUILD_TAG}"
                    mkdir -p artifacts
                    scp -q -r \
                        "${BUILD_USER}@${BUILD_SERVER}:${REMOTE_DIR}/desktop/dist" \
                        artifacts/desktop-dist
                '''
            }
        }
    }

    post {
        always {
            archiveArtifacts artifacts: 'artifacts/**', allowEmptyArchive: true
            junit allowEmptyResults: true, testResults: '**/test-results/*.xml'
            sh '''
                REMOTE_DIR="${REMOTE_ROOT}/dawg-${BUILD_TAG}"
                ssh -o BatchMode=yes "${BUILD_USER}@${BUILD_SERVER}" \
                    "rm -rf '${REMOTE_DIR}'" || true
            '''
        }
    }
}
