# Webhooker

Webserver para receber webhooks de repositórios git (até o momento, apenas suporta push e pullrequest do Bitbucket) e acionar uma pipeline que está descrita em um repositória a ser clonado.

## Variáveis de ambiente
- GIT_ROOT_FOLDER: Caminho absoluto onde os repositórios git serão configurados (ex.: /app/git)
- GIT_REPOS_TO_CLONE: Lista, separada por vírgula, de repositórios git para clonar dentro do GIT_ROOT_FOLDER (ex.: "git@bitbucket.org:example/example.git,git@bitbucket.org:example/example-2")
- PIPELINE_SCRIPT_PATH: Caminho, relativo ao GIT_ROOT_FOLDER, onde se encontra o arquivo de pipeline a ser chamado (ex.: example/pipeline/pipe.sh)
- COMMIT_MESSAGE_PREFIX_TO_IGNORE: Ignorar commits que suas mensagens começam com esse commit (ex.: "[X] ")
- COMMIT_AUTHOR_TO_IGNORE: Ignorar commits que seus autores sejam (ex.: "Pipeline <devops@example.com>")
- PUSH_TRIGGER_ONLY_IF_BRANCHES: Ignorar commits para branches de destino que não estejam incluída na lista, separada por vírgula (ex.: "main,master")