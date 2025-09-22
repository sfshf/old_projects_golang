export interface Site {
  id: number;
  name: string;
  admins: number[];
}

export interface Category {
  id: number;
  siteID: number;
  name: string;
  posts: number;
}

export interface Post {
  id: number;
  siteID: number;
  category: string;
  categoryID: number;
  title: string;
  postedAt: number;
  postedBy: number;
  postedByString: string;
  content: string;
  image: string;
  state: number;
  replies: number;
  views: number;
  activity: number;
  thumbups: number;
  thumbup: boolean;
}

export interface Comment {
  id: number;
  siteID: number;
  categoryID: number;
  postID: number;
  title: string;
  rootCommentID: number;
  postedAt: number;
  postedBy: number;
  postedByString: string;
  atWho: number;
  atWhoString: string;
  content: string;
  replies: number;
  thumbups: number;
  thumbup: boolean;
}

export interface Thumbup {
  id: number;
  siteID: number;
  categoryID: number;
  postID: number;
  title: string;
  commentID: number;
  type: string;
  postedAt: number;
  postdBy: string;
  postedByString: string;
  content: string;
}
