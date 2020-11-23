// const BundleAnalyzerPlugin = require("webpack-bundle-analyzer").BundleAnalyzerPlugin;
const webpack = require("webpack");

module.exports = {
  integrity: false,
  productionSourceMap: false,
  pwa: {
    name: "Errorly",
    themeColor: "#212529",
    msTileColor: "#212529",
    appleMobileWebAppCapable: "yes",
    appleMobileWebAppStatusBarStyle: "black-translucent",
    manifestOptions: {
      short_name: "Errorly",
      name: "Errorly",
      lang: "en",
      description:
        "Errorly is a service that allows you to automate data collection and handle them with ease in a kanban like style",
      background_color: "#ffffff",
      theme_color: "#212529",
      dir: "ltr",
      display: "standalone",
      orientation: "any",
      prefer_related_applications: false,
    },
    workboxOptions: {
      skipWaiting: true,
      clientsClaim: true,
    },
  },
  configureWebpack: {
    plugins: [
      // new BundleAnalyzerPlugin(),
      new webpack.ContextReplacementPlugin(/moment[/\\]locale$/, /en/),
    ],
    externals: {
      moment: "moment",
    },
  },
  devServer: {
    proxy: {
      "^/api": {
        methods: ["GET", "POST", "PUT", "HEAD"],
        target: "http://127.0.0.1:8001",
        ws: true,
        changeOrigin: true,
        withCredentials: true,
        secure: false,
      },
      "/login": {
        target: "http://127.0.0.1:8001",
        ws: true,
        changeOrigin: true,
        withCredentials: true,
        secure: false,
      },
      "/logout": {
        target: "http://127.0.0.1:8001",
        ws: true,
        changeOrigin: true,
        withCredentials: true,
        secure: false,
      },
      "/oauth2/callback": {
        target: "http://127.0.0.1:8001",
        ws: true,
        changeOrigin: true,
        withCredentials: true,
        secure: false,
      },
    },
  },
};
