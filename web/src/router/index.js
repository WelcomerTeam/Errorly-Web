import Vue from "vue";
import VueRouter from "vue-router";
import VueTimeAgo from "vue-timeago";
import { ToastPlugin } from "bootstrap-vue";

import Home from "../views/Home.vue";
import Projects from "../views/Projects.vue";
import CreateProject from "../views/CreateProject.vue";
import PageNotFound from "../views/PageNotFound.vue";

const Project = () =>
  import(/* webpackChunkName: "project" */ "../views/Project.vue");
const ProjectOverview = () =>
  import(/* webpackChunkName: "project" */ "../views/ProjectOverview.vue");
const ProjectIssues = () =>
  import(/* webpackChunkName: "project" */ "../views/ProjectIssues.vue");
const ProjectSettings = () =>
  import(/* webpackChunkName: "project" */ "../views/ProjectSettings.vue");
const CreateProjectIssue = () =>
  import(/* webpackChunkName: "project" */ "../views/CreateProjectIssue.vue");
const ViewProjectIssue = () =>
  import(/* webpackChunkName: "project" */ "../views/ViewProjectIssue.vue");

Vue.use(ToastPlugin);
Vue.use(VueRouter);
Vue.use(VueTimeAgo, {
  name: "Timeago",
  locale: "en",
});

const routes = [
  {
    path: "/",
    name: "Home",
    component: Home,
  },
  {
    path: "/projects",
    name: "Projects",
    component: Projects,
  },
  {
    path: "/project/create",
    name: "CreateProject",
    component: CreateProject,
  },
  {
    path: "/project/:id",
    component: Project,
    children: [
      {
        path: "",
        name: "ProjectOverview",
        component: ProjectOverview,
      },
      {
        path: "issues",
        name: "ProjectIssues",
        component: ProjectIssues,
      },
      {
        path: "settings",
        name: "ProjectSettings",
        component: ProjectSettings,
      },
      {
        path: "issue/create",
        component: CreateProjectIssue,
      },
      {
        path: "issue/:issueid",
        component: ViewProjectIssue,
      },
    ],
  },
  {
    path: "*",
    component: PageNotFound,
  },
];

// {
//   path: "/about",
//   name: "About",
//   // route level code-splitting
//   // this generates a separate chunk (about.[hash].js) for this route
//   // which is lazy-loaded when the route is visited.
//   component: () =>
//     import(/* webpackChunkName: "about" */ "../views/About.vue")
// }

const router = new VueRouter({
  mode: "history",
  base: process.env.BASE_URL,
  routes,
});

export default router;
