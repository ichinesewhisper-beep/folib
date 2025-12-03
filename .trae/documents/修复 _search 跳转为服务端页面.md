## 方案改进
- 使用绝对 URL 进行硬跳转：`location.href = location.origin + base_path + "/search"`，避免 SPA 识别为相对路径。
- 按钮改为原生 `<a>` 标签，且 `href` 使用绝对地址：`href={location.origin + base_path + "/search"}`。
- 文件修改：`OpenList-Frontend/src/pages/home/header/Header.tsx`，移除 `LinkWithBase`，引入 `base_path` 并拼接绝对 URL。

## 原理
- 绝对地址强制浏览器向服务端发起请求，前端路由不会拦截；即使 `base_path` 以 `/` 开头，`location.origin + base_path` 也生成正确的站点根绝对路径。

## 验证
- 点击放大镜按钮后整页加载服务端的 `/search` 页面；地址栏显示完整绝对 URL；不会出现“将 search 识别为文件夹”的问题。