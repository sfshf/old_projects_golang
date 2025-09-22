let password = process.env.NEXT_PUBLIC_PASSWORD ?? '';
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
  password = newPassword;
  if (typeof localStorage !== 'undefined' && password !== '') {
    localStorage.setItem('password', password);
  }
};

const shared = {
  getPassword,
  setPassword,
};

export default shared;
