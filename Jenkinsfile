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
                        "source /root/.nvm/nvm.sh && \
                         export PATH=/root/.nvm/versions/node/v26.5.0/bin:/usr/local/go/bin:\$PATH && \
                         export DAWG_CHROMIUM_EXECUTABLE_PATH=/root/.cache/ms-playwright/chromium-1193/chrome-linux/chrome && \
                         cd '${REMOTE_DIR}/engine' && \
                         go version && \
                         node --version && \
                         npm --version && \
                         rustc --version && \
                         cargo --version && \
                         npm ci && \
                         test -x /root/.cache/ms-playwright/chromium-1193/chrome-linux/chrome && \
                         gofmt -l . > /tmp/dawg-gofmt-files && \
                         test ! -s /tmp/dawg-gofmt-files && \
                         go vet ./... && \
                         go test ./... && \
                         go build ./cmd/dawg"
                    ssh -o BatchMode=yes "${BUILD_USER}@${BUILD_SERVER}" \
                        "source /root/.nvm/nvm.sh && \
                         export PATH=/root/.nvm/versions/node/v26.5.0/bin:/usr/local/go/bin:\$PATH && \
                         export DAWG_CHROMIUM_EXECUTABLE_PATH=/root/.cache/ms-playwright/chromium-1193/chrome-linux/chrome && \
                         cd '${REMOTE_DIR}/desktop' && \
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
