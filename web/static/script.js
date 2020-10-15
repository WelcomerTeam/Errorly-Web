// Install any plugins
Vue.use(VueChartJs);

Vue.component("pie-chart", {
    extends: VueChartJs.Pie,
    mixins: [VueChartJs.mixins.reactiveProp],
    props: ['chartData', 'options'],
    mounted() {
        this.renderChart(this.chartData, this.options)
    },
})

var vue;
var router;

function registerRouteEnter(routes, path, f) {
    routes.forEach(route => {
        if (route.path == path[0]) {
            if (path.length == 1) {
                route.component.beforeRouteEnter = f
                return
            } else {
                if (route.children?.length > 0 && path.length > 1) {
                    registerRouteEnter(route.children, path.slice(1), f)
                }
            }
        }
    })
}

axios.get("/api/dictionary").then(
    result => {
        routes = result.data.data.routes;

        registerRouteEnter(routes, ["/"], function (to, from, next) {
            console.log(to, from, next);
            next();
        })

        router = new VueRouter({
            routes,
        })

        vue = new Vue({
            router,
            data() {
                return {
                    hasError: false,
                    error: "",

                    userLoading: true,
                    userAuthenticated: false,
                    userProjects: [],
                    user: {},

                    projectFilter: "",
                }
            },
            mounted() {
                this.fetchMe();
            },
            methods: {
                fetchMe() {
                    axios.get("/api/me")
                        .then(result => {
                            data = result.data;
                            if (data.success) {
                                this.userAuthenticated = data.data.authenticated;
                                this.userProjects = data.data.projects;
                                this.user = data.data.user;
                            } else {
                                this.hasError = true;
                                this.error = data.error;
                            }
                        })
                        .catch(error => { this.hasError = true; this.error = error; })
                        .finally(() => { this.userLoading = false })
                }
            },
            computed: {
                filterProjects() {
                    if (this.projectFilter == "") { return this.userProjects }
                    return this.userProjects.filter(object => {
                        return object.name.toLowerCase().includes(this.projectFilter.toLowerCase())
                    })
                }
            },
            beforeRouteEnter(to, from, next) {
                console.log(to, from, next);
            },
        }).$mount("#app")
    }
)
