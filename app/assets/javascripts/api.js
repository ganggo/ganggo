var API = new RestClient("/api/v0", {
  contentType: 'application/x-www-form-urlencoded'}
);

API.res({
  posts: [
    'comments',
    'likes',
    'reshare'
  ],
  comments: [],
  notifications: [],
  people: [
    'profile',
    'aspects'
  ],
  aspects: [],
  oauth: [
    'tokens'
  ],
  users:[
    'streams'
  ],
});
