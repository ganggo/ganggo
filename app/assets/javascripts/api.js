var API = new RestClient("/api/v0")
API.res({
  posts: [
    'comments',
    'likes'
  ],
  people: ['profile']
});
