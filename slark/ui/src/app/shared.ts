let user = process.env.NEXT_PUBLIC_USER ?? '';
let password = process.env.NEXT_PUBLIC_PASSWORD ?? '';
const baseAPIURL = process.env.NEXT_PUBLIC_SERVER_BASE_URL;
let loadedFromLocalStorage = false;

const getPassword = () => {
  // try to load from local storage
  if (
    password === '' &&
    !loadedFromLocalStorage &&
    typeof localStorage !== 'undefined'
  ) {
    loadedFromLocalStorage = true;
    const storedPassword = localStorage.getItem('password');
    if (storedPassword) {
      password = storedPassword;
    }
  }
  return password;
};

const setPassword = (newPassword: string) => {
  // console.log("setPassword", newPassword);
  password = newPassword;
  if (typeof localStorage !== 'undefined' && password !== '') {
    localStorage.setItem('password', password);
  }
};

const shared = {
  getPassword,
  setPassword,
  baseAPIURL,
};

export default shared;
