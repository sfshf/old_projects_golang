/**
 * Sample React Native App
 * https://github.com/facebook/react-native
 *
 * @format
 */

import React from 'react';
import { useTranslation } from 'react-i18next';
import {
  makeStyles,
  Text,
  Button,
  Overlay,
  useTheme,
  Switch,
} from '@rneui/themed';
import { Pressable, View } from 'react-native';
import moment from 'moment';
import WheelPicker from '@quidone/react-native-wheel-picker';
import WheelPickerFeedback from '@quidone/react-native-wheel-picker-feedback';

export interface Time {
  hour: number;
  minute: number;
  second: number;
}

export interface Date {
  year: number;
  month: number;
  day: number;
}

type DateAction =
  | { type: 'set'; value: Date }
  | { type: 'setYear'; value: Date['year'] }
  | { type: 'setMonth'; value: Date['month'] }
  | {
      type: 'setDay';
      value: Date['day'];
    };

const initDate: Date = {
  year: moment().year(),
  month: moment().month(),
  day: moment().day(),
};

const dateReducer = (state: Date, action: DateAction) => {
  switch (action.type) {
    case 'set':
      return { ...action.value };
    case 'setYear':
      if (state.year !== action.value) {
        return { ...state, year: action.value };
      }
      break;
    case 'setMonth':
      if (state.month !== action.value) {
        return { ...state, month: action.value };
      }
      break;
    case 'setDay':
      if (state.day !== action.value) {
        return { ...state, day: action.value };
      }
      break;
  }
  return state;
};

const arrayRange = (start: number, stop: number, step: number): number[] =>
  Array.from(
    { length: (stop - start) / step + 1 },
    (value, index) => start + index * step,
  );

type DatePickerProps = {
  visible: boolean;
  setVisible: (visible: boolean) => void;
  value: string;
  setValue: (value: string) => void;
};

function DatePicker({
  visible,
  setVisible,
  value,
  setValue,
}: DatePickerProps): React.JSX.Element {
  const { t } = useTranslation();
  const styles = useStyles();
  const { theme } = useTheme();

  const toggleOverlay = () => {
    setVisible(!visible);
  };

  const isByMonth = () => {
    if (value) {
      const date = value.split('-');
      if (date.length === 3) {
        return false;
      } else if (date.length === 2) {
        return true;
      }
    }
    return false;
  };

  const [byMonth, setByMonth] = React.useState(isByMonth());

  const parseValue = () => {
    if (value) {
      const date = value.split('-');
      if (date.length === 3) {
        return {
          year: parseInt(date[0]),
          month: parseInt(date[1]),
          day: parseInt(date[2]),
        };
      } else if (date.length === 2) {
        return {
          year: parseInt(date[0]),
          month: parseInt(date[1]),
          day: 1,
        };
      }
    }
    return initDate;
  };
  const [date, dispatchDate] = React.useReducer(dateReducer, parseValue());

  React.useEffect(() => {
    dispatchDate({ type: 'set', value: parseValue() });
    setByMonth(isByMonth());
  }, [value]);

  const [years, setYears] = React.useState<number[]>(
    arrayRange(1900, initDate.year + 100, 1),
  );
  const [months, setMonths] = React.useState<number[]>(arrayRange(1, 12, 1));
  const [days, setDays] = React.useState<number[]>(
    arrayRange(
      1,
      moment(date.year + '-' + date.month, 'YYYY-MM').daysInMonth(),
      1,
    ),
  );

  const onPressTopClose = () => {
    setVisible(false);
  };

  const onPressUse = () => {
    if (!byMonth) {
      setValue(date.year + '-' + date.month + '-' + date.day);
    } else {
      setValue(date.year + '-' + date.month);
    }
    setVisible(false);
  };

  return (
    <Overlay
      fullScreen
      overlayStyle={styles.container}
      isVisible={visible}
      onBackdropPress={toggleOverlay}>
      <View style={styles.body}>
        <View style={styles.topline}>
          <Pressable style={styles.pressable} onPress={onPressTopClose} />
        </View>
        <View style={[styles.row, { justifyContent: 'flex-start' }]}>
          <Text style={styles.fieldLabel}>
            {t('records.newRecord.datePicker.byMonth')}
          </Text>
          <Switch
            trackColor={{
              false: theme.colors.grey3,
              true: theme.colors.primary,
            }}
            value={byMonth}
            onValueChange={value => {
              setByMonth(value as boolean);
            }}
          />
        </View>
        <View
          style={[
            styles.row,
            { marginVertical: 4, justifyContent: 'space-around' },
          ]}>
          <Text style={styles.fieldLabel}>
            {t('records.newRecord.datePicker.yearLabel')}
          </Text>
          <Text style={styles.fieldLabel}>
            {t('records.newRecord.datePicker.monthLabel')}
          </Text>
          {!byMonth && (
            <Text style={styles.fieldLabel}>
              {t('records.newRecord.datePicker.dayLabel')}
            </Text>
          )}
        </View>
        <View
          style={[
            styles.row,
            { marginVertical: 4, justifyContent: 'space-around' },
          ]}>
          <WheelPicker
            width={80}
            onValueChanging={() => {
              WheelPickerFeedback.triggerSoundAndImpact();
            }}
            data={years.map(index => ({
              value: index,
              label: index.toString(),
            }))}
            value={date.year}
            onValueChanged={({ item: { value } }) => {
              dispatchDate({ type: 'setYear', value: value });
              setDays(
                arrayRange(
                  1,
                  moment(value + '-' + date.month, 'YYYY-MM').daysInMonth(),
                  1,
                ),
              );
            }}
          />
          <WheelPicker
            width={80}
            onValueChanging={() => {
              WheelPickerFeedback.triggerSoundAndImpact();
            }}
            data={months.map(index => ({
              value: index,
              label: index.toString(),
            }))}
            value={date.month}
            onValueChanged={({ item: { value } }) => {
              dispatchDate({ type: 'setMonth', value: value });
              setDays(
                arrayRange(
                  1,
                  moment(date.year + '-' + value, 'YYYY-MM').daysInMonth(),
                  1,
                ),
              );
            }}
          />
          {!byMonth && (
            <WheelPicker
              width={80}
              onValueChanging={() => {
                WheelPickerFeedback.triggerSoundAndImpact();
              }}
              data={days.map(index => ({
                value: index,
                label: index.toString(),
              }))}
              value={date.day}
              onValueChanged={({ item: { value } }) =>
                dispatchDate({ type: 'setDay', value: value })
              }
            />
          )}
        </View>
        <View style={styles.row}>
          <Button
            type="solid"
            radius={8}
            color={theme.colors.primary}
            containerStyle={styles.btnContainer}
            titleStyle={styles.useBtnTitle}
            title={t('records.newRecord.datePicker.useBtn')}
            onPress={onPressUse}
          />
        </View>
      </View>
    </Overlay>
  );
}

const useStyles = makeStyles(theme => ({
  container: {
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    marginTop: '160%',
  },
  body: {
    marginHorizontal: 8,
  },
  topline: {
    height: 10,
    alignItems: 'center',
  },
  pressable: {
    height: 4,
    width: 40,
    marginVertical: 1,
    borderRadius: 2,
    backgroundColor: theme.colors.surface,
  },
  row: {
    flexDirection: 'row',
    marginVertical: 16,
    marginHorizontal: 8,
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  btnContainer: { width: '100%' },
  useBtnTitle: { fontSize: 20 },
  pswdText: {
    marginVertical: 8,
    marginHorizontal: 8,
    borderWidth: 2,
    borderRadius: 8,
    borderColor: theme.colors.black,
    fontSize: 20,
    height: 40,
  },
  fieldLabel: {
    fontSize: 20,
    fontWeight: 'bold',
    color: theme.colors.black,
    alignContent: 'center',
  },
  charLengthSlider: { width: '50%' },
}));

export default DatePicker;
