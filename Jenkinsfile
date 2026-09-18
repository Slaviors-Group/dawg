pipeline {
    agent any

    options {
        buildDiscarder(logRotator(numToKeepStr: '20', artifactNumToKeepStr: '2'))
        disableConcurrentBuilds(abortPrevious: true)
        skipDefaultCheckout(true)
        timestamps()
        timeout(time: 150, unit: 'MINUTES')
    }

    environment {
        BUILD_HOST = 'x220-builder'
        REMOTE_ROOT = '/home/dawg-builder/jenkins-workspaces/dawg'
        DAWG_CI_NODE_VERSION = '24.18.0'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
                script {
                    env.GIT_SHA = sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
                    def safeBuildTag = env.BUILD_TAG.replaceAll(/[^A-Za-z0-9_.-]/, '_')
                    env.REMOTE_DIR = "${env.REMOTE_ROOT}/${safeBuildTag}"
                }
            }
        }

        stage('Builder preflight') {
            steps {
                sh '''
                    set -eu
                    ssh -o BatchMode=yes "$BUILD_HOST" "
                        set -eu
                        test \"\$(id -un)\" = dawg-builder
                        if sudo -n true >/dev/null 2>&1; then
                            echo 'The CI account unexpectedly has sudo access.' >&2
                            exit 1
                        fi
                        command -v /usr/local/go/bin/go >/dev/null
                        command -v \"\$HOME/.cargo/bin/cargo\" >/dev/null
                        command -v patchelf >/dev/null
                        pkg-config --exists librsvg-2.0
                        available_kb=\$(df -Pk \"\$HOME\" | awk 'NR == 2 { print \$4 }')
                        if [ \"\$available_kb\" -lt 15728640 ]; then
                            echo 'The build server needs at least 15 GiB of free disk space.' >&2
                            exit 1
                        fi
                    "
                '''
            }
        }

        stage('Transfer source') {
            steps {
                sh '''
                    set -eu
                    ssh -o BatchMode=yes "$BUILD_HOST" "install -d -m 0755 '$REMOTE_DIR'"
                    git archive --format=tar HEAD |
                        ssh -o BatchMode=yes "$BUILD_HOST" "tar -xf - -C '$REMOTE_DIR'"
                '''
            }
        }

        stage('Validate') {
            steps {
                sh '''
                    set -eu
                    ssh -o BatchMode=yes "$BUILD_HOST" "
                        set -eu
                        cd '$REMOTE_DIR'
                        DAWG_CI_NODE_VERSION='$DAWG_CI_NODE_VERSION' \
                            ./ci/linux/with-toolchain.sh ./ci/linux/validate.sh
                    "
                '''
            }
        }

        stage('Build Linux AppImage') {
            when {
                expression {
                    def ref = env.BRANCH_NAME ?: env.GIT_BRANCH ?: ''
                    return env.TAG_NAME || ref == 'main' || ref == 'staging' ||
                        ref ==~ /(?:origin\/)?v.*/
                }
            }
            steps {
                sh '''
                    set -eu
                    ssh -o BatchMode=yes "$BUILD_HOST" "
                        set -eu
                        cd '$REMOTE_DIR'
                        DAWG_CI_NODE_VERSION='$DAWG_CI_NODE_VERSION' \
                            DAWG_CI_GIT_COMMIT='$GIT_SHA' \
                            DAWG_BUNDLE_CACHE_DIR='/home/dawg-builder/.cache/dawg-ci/bundle-downloads' \
                            DAWG_BUNDLE_BROWSER_CACHE_DIR='/home/dawg-builder/.cache/dawg-ci/playwright-browsers' \
                            ./ci/linux/with-toolchain.sh ./ci/linux/build.sh
                        DAWG_CI_NODE_VERSION='$DAWG_CI_NODE_VERSION' \
                            DAWG_CI_GIT_COMMIT='$GIT_SHA' \
                            ./ci/linux/with-toolchain.sh ./ci/linux/verify-appimage.sh
                    "
                '''
            }
        }

        stage('Collect artifacts') {
            when {
                expression {
                    def ref = env.BRANCH_NAME ?: env.GIT_BRANCH ?: ''
                    return env.TAG_NAME || ref == 'main' || ref == 'staging' ||
                        ref ==~ /(?:origin\/)?v.*/
                }
            }
            steps {
                retry(2) {
                    sh '''
                        ./ci/linux/collect-artifacts.sh \
                            "$BUILD_HOST" "$REMOTE_DIR/.ci-artifacts" artifacts
                    '''
                }
                archiveArtifacts artifacts: 'artifacts/**', fingerprint: true
            }
        }
    }

    post {
        always {
            sh '''
                set +e
                if [ -n "${REMOTE_DIR:-}" ]; then
                    ssh -o BatchMode=yes "$BUILD_HOST" "
                        case '$REMOTE_DIR' in
                            '$REMOTE_ROOT'/*) rm -rf -- '$REMOTE_DIR' ;;
                            *) echo 'Refusing unsafe remote cleanup path.' >&2; exit 2 ;;
                        esac
                    "
                fi
            '''
            deleteDir()
        }
    }
}
