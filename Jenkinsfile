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
                    cargo_status=0
                    ssh -o BatchMode=yes "${BUILD_USER}@${BUILD_SERVER}" \
                        "source /root/.nvm/nvm.sh && \
                         export PATH=/root/.nvm/versions/node/v26.5.0/bin:/usr/local/go/bin:/root/.cargo/bin:\$PATH && \
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
                         export PATH=/root/.nvm/versions/node/v26.5.0/bin:/usr/local/go/bin:/root/.cargo/bin:\$PATH && \
                         export DAWG_CHROMIUM_EXECUTABLE_PATH=/root/.cache/ms-playwright/chromium-1193/chrome-linux/chrome && \
                         cd '${REMOTE_DIR}/desktop' && \
                         npm ci && \
                         npx --no-install biome check . && \
                         npm run build && \
                         chmod +x build-bundle.sh && \
                         ./build-bundle.sh --skip-tauri"

                    cargo_status=0
                    ssh -o BatchMode=yes "${BUILD_USER}@${BUILD_SERVER}" \
                        bash -s -- "${REMOTE_DIR}" <<'REMOTE_CARGO_SCRIPT' || cargo_status=$?
                    set +e
                    REMOTE_DIR="$1"
                    source /root/.nvm/nvm.sh
                    export PATH=/root/.nvm/versions/node/v26.5.0/bin:/usr/local/go/bin:/root/.cargo/bin:$PATH
                    export DAWG_CHROMIUM_EXECUTABLE_PATH=/root/.cache/ms-playwright/chromium-1193/chrome-linux/chrome
                    cd "$REMOTE_DIR/desktop/src-tauri"
                    export CARGO_TARGET_DIR="$REMOTE_DIR/cargo-target"
                    mkdir -p "$CARGO_TARGET_DIR"

                    cargo_status=0
                    CARGO_BUILD_JOBS=1 cargo check --locked --verbose || cargo_status=$?
                    if [ "$cargo_status" -ne 0 ]; then
                        {
                            echo '--- working directory ---'
                            pwd
                            echo '--- toolchain ---'
                            rustc -vV
                            cargo -vV
                            echo '--- filesystem ---'
                            df -h "$REMOTE_DIR" /tmp
                            df -i "$REMOTE_DIR" /tmp
                            echo '--- target directories ---'
                            find "$CARGO_TARGET_DIR" -maxdepth 5 -type d -print | sort
                            echo '--- proc-macro2 output parent ---'
                            find "$CARGO_TARGET_DIR/debug/build" -path '*proc-macro2*' -maxdepth 4 -print 2>/dev/null
                            echo '--- build output parents ---'
                            find "$CARGO_TARGET_DIR/debug/build" -mindepth 2 -maxdepth 2 -type d -printf '%p\n' 2>/dev/null | sort
                            echo '--- target directory metadata ---'
                            stat "$CARGO_TARGET_DIR" "$CARGO_TARGET_DIR/debug" "$CARGO_TARGET_DIR/debug/build" 2>&1
                        } > "$REMOTE_DIR/cargo-diagnostics.txt" 2>&1
                    fi
                    exit "$cargo_status"
REMOTE_CARGO_SCRIPT

                    mkdir -p artifacts
                    scp -q \
                        "${BUILD_USER}@${BUILD_SERVER}:${REMOTE_DIR}/cargo-diagnostics.txt" \
                        artifacts/cargo-diagnostics.txt || true
                    if [ -f artifacts/cargo-diagnostics.txt ]; then
                        cat artifacts/cargo-diagnostics.txt
                    fi
                    exit "$cargo_status"
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
