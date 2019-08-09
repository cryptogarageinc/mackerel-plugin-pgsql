# settlenet-mackerel-pgsql
mackerel plugin for executing sql on postgres

# How to release for mkr install
1. `$ make setup`
1. `$ git tag v0.19.1` (タグ名は適宜置き換えること)
1. `$ GITHUB_TOKEN=... script/release.sh`
