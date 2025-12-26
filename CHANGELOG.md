> 出于安全考虑，目前仅支持 Docker/Compose 部署方式。[镜像地址](https://github.com/engigu/baihu-panel/pkgs/container/baihu)

## 快速部署

```bash
docker run -d \
  --name baihu \
  -p 8052:8052 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/envs:/app/envs \
  -e TZ=Asia/Shanghai \
  -e BH_DB_TYPE=sqlite \
  -e BH_DB_PATH=/app/data/ql.db \
  -e BH_DB_TABLE_PREFIX=baihu_ \
  --restart unless-stopped \
  ghcr.io/engigu/baihu:latest
```

启动后访问：http://localhost:8052

默认账号：`admin` / `123456`


