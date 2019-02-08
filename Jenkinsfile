library identifier: "pipeline-library@master",
retriever: modernSCM(
  [
    $class: "GitSCMSource",
    remote: "https://github.com/redhat-cop/pipeline-library.git"
  ]
)

openshift.withCluster() {
  env.NAMESPACE = openshift.project()
  env.APP_NAME = "cert-operator"
  ARTIFACT_DIRECTORY = "build/bin"
}

pipeline {

  agent {
    label 'jenkins-slave-golang'
  }

  stages {
  	stage('Git Checkout') {
      steps {
        git url: "${SOURCE_REPOSITORY_URL}", branch: "${SOURCE_REPOSITORY_REF}"
      }
    }

    stage('Build Cert Operator') {
      steps {
       	sh """
          export GOPATH=${WORKSPACE}
          cd ..
          cp -R cert-operator/ cert-operator/src/github.com/redhat-cop/cert-operator/
          cd cert-operator/src/github.com/redhat-cop/cert-operator/
          ls -al
          pwd
          ./build.sh
        """
      }
  	}

  	stage('Build Image') {
      steps {
       	binaryBuild(projectName: "${NAMESPACE}", buildConfigName: "${APP_NAME}", artifactsDirectoryName: "${ARTIFACT_DIRECTORY}")
      }
    }
  }
}