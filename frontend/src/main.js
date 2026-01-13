import { createApp } from "vue";
import ElementPlus from "element-plus";
import "element-plus/dist/index.css";
import App from "./App.vue";
import "./assets/base.css";
import "xterm/css/xterm.css";

createApp(App).use(ElementPlus).mount("#app");
