@echo off
git add .
git commit -m "feat: redis实现定时写入视频浏览量，视频热度计算与排序"
git push

echo "github push successfully!"