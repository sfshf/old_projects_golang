package consts

// accent
type AudioAccent string

const (
	US AudioAccent = "us"
	UK AudioAccent = "uk"
)

// voice
type AudioVoice string

const (
	// Joey AudioVoiceUS = "Joey" Matthew
	// us
	US_Matthew AudioVoice = "Matthew" // male
	US_Joanna  AudioVoice = "Joanna"  // female
	// uk
	UK_Brian AudioVoice = "Brian" // male
	UK_Amy   AudioVoice = "Amy"   // femal
)

// Speed
type AudioSpeed float32

const (
	Speed1_0 AudioSpeed = 1.0
	Speed1_1 AudioSpeed = 1.1
	Speed1_2 AudioSpeed = 1.2
	Speed0_9 AudioSpeed = 0.9
	Speed0_8 AudioSpeed = 0.8
)
