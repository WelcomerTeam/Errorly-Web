import Vue from "vue";
import VueRouter from "vue-router";
import VueTimeAgo from "vue-timeago";

import PageNotFound from "../views/PageNotFound.vue";
import Home from "../views/Home.vue";
import Projects from "../views/Projects.vue";
import CreateProject from "../views/CreateProject.vue";

import Project from "../views/Project.vue";
import ProjectOverview from "../views/ProjectOverview.vue";
import ProjectIssues from "../views/ProjectIssues.vue";
import ProjectSettings from "../views/ProjectSettings.vue";

import CreateProjectIssue from "../views/CreateProjectIssue.vue";
import ViewProjectIssue from "../views/ViewProjectIssue.vue";

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
