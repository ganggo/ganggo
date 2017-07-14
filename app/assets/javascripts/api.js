var API = new RestClient("/api/v0", {
  contentType: 'application/x-www-form-urlencoded'}
);

API.res({
  posts: [
    'comments',
    'likes'
  ],
  people: [
    'profile',
    'aspects'
  ],
  aspects: []
});
