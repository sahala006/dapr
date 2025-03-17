# 使用官方的Python镜像作为基础镜像
FROM python:3.7.5
 
# 设置工作目录
WORKDIR /root/oms

# 将项目代码复制到容器中
COPY . .

# 安装依赖
RUN pip install --no-cache-dir -r requirements.txt

# 暴露端口
EXPOSE 8100
 
# 运行Django项目
CMD ["python", "manage.py", "runserver", "0.0.0.0:8100"]
