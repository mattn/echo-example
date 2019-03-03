var app = new Vue({
  el: '#app',
  data: {
    comments: [],
    newName: '',
    newText: '',
  },
  created: () => {
    axios.get('/api/comments')
      .then((response) => {
        console.log(response);
		  alert(1);
        this.comments = response.data.items || [];
      })
      .catch((error) => {
        console.log(error);
      });
  },
  methods: {
    addComment: () => {
      let params = new URLSearchParams();
      params.append('name', this.newName);
      params.append('text', this.newText);
      axios.post('/api/comments', params)
        .then((response) => {
          console.log(response)
          vue.$forceUpdate()
        })
        .catch((error) => console.log(error))
    }
  }
})
