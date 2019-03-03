const app = new Vue({
  el: '#app',
  data: {
    comments: [],
    name: '',
    text: '',
  },
  created() { this.update() },
  methods: {
    add: () => {
      const payload = {'name': app.name, 'text': app.text}
      axios.post('/api/comments', payload)
        .then(() => {
          app.name = ''
          app.text = ''
          app.update()
        })
        .catch((err) => {
          alert(err.response.data.error)
        })
    },
    update: () => {
      axios.get('/api/comments')
        .then((response) => app.comments = response.data || [])
        .catch((error) => console.log(error));
    }
  }
})
